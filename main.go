package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"louie-web-administrator/admin"
	"louie-web-administrator/configuration"
	"louie-web-administrator/dashboard"
	"louie-web-administrator/louie_kafka"
	"louie-web-administrator/repository"
	"louie-web-administrator/service"
	"louie-web-administrator/websocket"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/unrolled/secure"
)

func main() {

	ctx := context.Background()

	secureMiddleware := secure.New(secure.Options{
		AllowedHosts:         []string{".*"},
		AllowedHostsAreRegex: true,
		HostsProxyHeaders:    []string{"X-Forwarded-Host"},
		SSLRedirect:          false,
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
	})

	listenAddr := ":5000"
	log.Printf("using Port: %s\n", listenAddr)

	// --- fill configuration with environments ---
	cfg := processConfiguration()
	// ---

	// --- init mongo db ---
	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      cfg.Database.DatabaseUser,
		Password:      cfg.Database.DatabasePassword,
	}
	clientOpts := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%d",
			cfg.Database.DatabaseServer,
			cfg.Database.DatabasePort)).
		SetAuth(credential)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	// ---

	// --- init repositories ---
	userRepository := repository.NewUserRepo(ctx, client, cfg.Database.DatabaseName)
	gameRepository := repository.NewGameRepository(ctx, client, cfg.Database.DatabaseName)
	// ---

	// --- init channels ---
	kafkaGameEventsChannel := make(chan kafka.Message)
	kafkaTechnicalEventsChannel := make(chan kafka.Message)
	kafkaQuitChannel := make(chan bool)
	dashboardChannel := make(chan *websocket.DashboardSignal, 100)
	adminUiChannel := make(chan websocket.AdminUiEvent, 10)
	// ---

	// --- init kafka consumer ---
	setupKafkaConsumer(ctx, cfg, kafkaGameEventsChannel, kafkaTechnicalEventsChannel, kafkaQuitChannel)
	// ---

	// --- init kafka producer ---
	kafkaProducer := louie_kafka.KafkaProducer{
		ServerAddress: fmt.Sprintf("%s:%s", cfg.Kafka.Server, cfg.Kafka.Port),
		Topic:         cfg.Kafka.LouieEventTopic,
	}
	// ---

	// --- init services ---
	userService := &service.UserSer{UserRepository: userRepository}
	gameService := &service.GameSer{
		UserRepository: userRepository,
		GameRepository: gameRepository,
		KafkaProducer:  kafkaProducer,
	}
	// ---

	// --- init websockets ---
	adminUiWebsocket := websocket.InitAdminUiWebsocket(adminUiChannel)

	dashboardWebsocket := websocket.InitGameDashboardSocket(
		dashboardChannel,
		gameService.GetCurrentDashboardState,
	)
	// ---

	// --- init technical event handler ---
	technicalEventHandler := service.RunTechnicalEventHandler(kafkaTechnicalEventsChannel, adminUiChannel, kafkaProducer)
	// ---

	// --- init state changer ---
	stateChanger := service.GameStateChecker{UserService: userService, GameService: gameService, GameDashboardSocket: *dashboardWebsocket, AdminUiSocket: *adminUiWebsocket}
	stateChanger.RunGameStateChecker(kafkaGameEventsChannel)
	// ---

	// --- init louki ki user ---
	userService.InitOrRefreshLouki()
	// ---

	// --- init admin event service ---
	adminEventService := service.InitAdminEventService(userService, gameService, adminUiWebsocket)
	// ---

	// --- init controller routes ---
	router := setupRoutes(userService, gameService, dashboardWebsocket, adminUiWebsocket, technicalEventHandler, adminEventService)

	server := &http.Server{
		Addr: listenAddr,
		Handler: handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"OPTIONS", "GET", "POST", "PUT"}),
			handlers.AllowedHeaders([]string{"Content-Type", "X-Auth-Token"}),
		)(secureMiddleware.Handler(handlers.CompressHandler(handlers.RecoveryHandler()(router)))),
	}
	// ---

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panicf("listen: %s\n", err)
		}
	}()

	log.Print("server started")
	<-done

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer func() {
		log.Print("stopping server")
		cancel()

		log.Printf("stopping kafka consumer")
		kafkaQuitChannel <- true
		close(kafkaGameEventsChannel)
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server Shutdown Failed: %+v", err)
	}
	log.Print("server exited")
}

func setupKafkaConsumer(
	ctx context.Context,
	config *configuration.Config,
	gameEventsChannel chan kafka.Message,
	technicalEventsChannel chan kafka.Message,
	quit chan bool) {

	consumerConfig := louie_kafka.ConsumerConfig{
		Quit:            quit,
		GameEvents:      gameEventsChannel,
		TechnicalEvents: technicalEventsChannel,
		ServerAddress: fmt.Sprintf("%s:%s",
			config.Kafka.Server,
			config.Kafka.Port,
		),
		Topic: config.Kafka.LouieEventTopic,
	}

	go consumerConfig.StartConsumer(ctx)
}

func setupRoutes(
	userService *service.UserSer,
	gameService *service.GameSer,
	gameDashboardSocket *websocket.GameDashboardSocket,
	adminUiWebsocket *websocket.AdminUiWebsocket,
	technicalEventHandler *service.TechnicalEventHandler,
	adminEventService *service.AdminEventService) *mux.Router {

	router := mux.NewRouter()

	router.
		HandleFunc("/", admin.Main(userService, gameService)).
		Methods("GET")

	router.
		HandleFunc("/confirm", admin.ConfirmSideChange(technicalEventHandler)).
		Methods("POST")

	router.
		HandleFunc("/user/filter", admin.Filter(userService)).
		Methods("POST").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	router.
		HandleFunc("/user/wait", admin.Wait(userService, adminEventService)).
		Methods("PUT").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	router.
		HandleFunc("/user/state", admin.State(userService, adminEventService)).
		Methods("PUT").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	router.
		HandleFunc("/user/position", admin.Position(userService)).
		Methods("PUT").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	router.
		HandleFunc("/user/position/shuffle", admin.Shuffle(userService)).
		Methods("PUT").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	router.
		HandleFunc("/user/{page:[0-9]+}", admin.Page(userService)).
		Methods("GET")

	router.
		HandleFunc("/game", admin.AnnounceGame(userService, gameService)).
		Methods("POST").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	router.
		HandleFunc("/game", admin.RemoveGame(userService, gameService, gameDashboardSocket)).
		Methods("PUT").
		Headers("Content-Type", "application/x-www-form-urlencoded")

	abs, err := filepath.Abs("./admin/static")

	if err != nil {
		log.Fatalf("can not find static files: %s", err)
	}

	fs := http.FileServer(http.Dir(abs))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	router.HandleFunc("/ws", adminUiWebsocket.AdminUiWebsocketEndpoint())

	// --- User-Ui Dashboard (Angular App) ---
	router.
		HandleFunc("/user", dashboard.Post(userService)).
		Methods("POST").
		Headers("Content-Type", "application/json")

	router.HandleFunc("/ws/game", gameDashboardSocket.GameDashboardWebsocketEndpoint())

	return router
}

func processConfiguration() *configuration.Config {
	cfg := configuration.Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
