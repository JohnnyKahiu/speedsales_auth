package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/JohnnyKahiu/speedsales_login/api"
	"github.com/JohnnyKahiu/speedsales_login/pkg/database"
	"github.com/JohnnyKahiu/speedsales_login/pkg/grpc"
	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
	"github.com/JohnnyKahiu/speedsales_login/pkg/variables"
	"github.com/joho/godotenv"
)

type dbConf struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	DbName   string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ConfigFile struct {
	Branch         string   `json:"branch"`
	Listen         string   `json:"listen"`
	Port           string   `json:"port"`
	MasterAddr     string   `json:"master_addr"`
	MirrorAddr     string   `json:"Mirror_addr"`
	MirrorPort     string   `json:"mirror_port"`
	ScServer       string   `json:"sc_server"`
	ScPort         int      `json:"sc_port"`
	ServerID       int64    `json:"server_id"`
	ServerName     string   `json:"server_name"`
	ServerBranches []string `json:"server_branches"`
	StockBranch    string   `json:"stock_branch"`
	RemoteSvrs     []string `json:"remote_servers"`
	MasterUrl      string   `json:"master_url"`
	CompanyName    string   `json:"company_name"`
	EtrSocket      string   `json:"etr_socket"`
	EtrType        string   `json:"etr_type"`
	ProductionDisp bool     `json:"production_disp"`
	IndustryMode   string   `json:"industrimode"`
}

func getRunningIPAddress() string {
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// os.Stdout.WriteString(ipnet.IP.String() + "\n")
				// fmt.Printf("\n%v", ipnet.IP.String())
				return ipnet.IP.String()
			}
		}
	}
	return "0.0.0.0"
}

func (arg *ConfigFile) readConfFile() error {
	file := variables.Fpath + "/config.json"
	fmt.Println("files =", file)
	jsonFile, err := os.Open(file)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return err
	}
	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &arg)

	return nil
}

func initTables() error {
	return users.GenUsersTables()
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	var err error
	// fetch file path from argument
	var serverType, cache string

	flag.StringVar(&variables.Fpath, "path", "", "local files path")
	flag.StringVar(&cache, "cache", "", "cache")
	flag.StringVar(&serverType, "st", "", "server type")

	// get tls certificates
	certFile := flag.String("certfile", "cert.pem", "certificate PEM file")
	keyFile := flag.String("keyfile", "key.pem", "key PEM file")

	isTLS := flag.Bool("tls", false, "enable tls")
	initDB := flag.Bool("initDB", false, "init db")

	flag.Parse()

	// enable environment files
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("server type = ", serverType)
	if variables.Fpath == "" {
		variables.Fpath, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
	}

	// get configuration files
	/*
		configs := ConfigFile{}
		err = configs.readConfFile()
		if err != nil {
			log.Println("failed to get config files.   err = ", err)
		}
	*/

	conf := database.DBConf{
		Server: os.Getenv("DB_HOST"),
		Port:   os.Getenv("DB_PORT"),
		DbName: os.Getenv("DB_NAME"),
	}
	// make a postgresql database connection
	database.PgPool, err = conf.NewPgPool()
	if err != nil {
		log.Fatalln("\t failed to connect Postgres Pool.    err =", err)
	}
	defer database.PgPool.Close()

	fmt.Println("init database = ", *initDB)
	if *initDB {
		err = initTables()
	}

	address := os.Getenv("LISTEN_ADDRESS")
	if address == "*" {
		address = "0.0.0.0"
	}
	if address == "card" {
		address = getRunningIPAddress()
	}

	go func() {
		fmt.Printf("\t grpc://%s:%s\n", address, os.Getenv("GRPC_SERVER_PORT"))
		if err := grpc.NewServer(address + ":" + os.Getenv("GRPC_SERVER_PORT")); err != nil {
			log.Fatalf("error    failed to start a new grpc server")
		}
	}()

	port := os.Getenv("PORT")

	r := api.NewRouter()
	if *isTLS {
		fmt.Printf("\thttps://%v:%v\n", address, port)
		srv := &http.Server{
			Addr:    address + ":" + port,
			Handler: r,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS13,
				PreferServerCipherSuites: true,
			},
		}
		err = srv.ListenAndServeTLS(*certFile, *keyFile)
		if err != nil {
			log.Fatal("failed to start tls server    err =", err)
		}
	} else {
		fmt.Printf("\thttp://%v:%v\n", address, port)
		// http.ListenAndServeTLS(address+":"+port, "localhost.crt", "localhost.key", r)

		http.ListenAndServe(address+":"+port, r)
	}
}
