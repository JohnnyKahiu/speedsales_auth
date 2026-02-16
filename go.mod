module github.com/JohnnyKahiu/speedsales_login

go 1.25.4

require (
	github.com/JohnnyKahiu/speed_sales_proto v0.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.8.0
	github.com/joho/godotenv v1.5.1
	golang.org/x/crypto v0.46.0
	google.golang.org/grpc v1.62.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace github.com/JohnnyKahiu/speed_sales_proto => ../proto/gen/github.com/JohnnyKahiu/speed_sales_proto
