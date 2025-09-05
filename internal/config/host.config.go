package config

import (
	"github.com/Ermi9s/Anubis/internal/model"
	"github.com/Ermi9s/Anubis/internal/repository"
	"github.com/Ermi9s/Anubis/internal/rpcserver"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"gopkg.in/yaml.v3"
)



func loadConfig(path string) (*model.Configuration, error) {
    file, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg model.Configuration
    if err := yaml.Unmarshal(file, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}



func selectQueue(cfg *model.Configuration, repository *repository.Repository) {
	if(cfg.RabbitMQ != nil) {
		go StartRabbitMQConsumer(cfg, repository)	
	}
	// add others when its time 

}

func selecteDatabase(cfg *model.Configuration) string {
	if(cfg.Postgres != nil) {
		return "postgres"
	}
	//other databases 

	return ""
}


func startRpc(repository *repository.Repository) {
	rmi := rpcserver.NewRpcServer(repository)
	rpc.Register(rmi)

	listener, err := net.Listen("tcp", ":8080") 
	if err != nil {
		fmt.Println("[Anubis] Error listening:", err)
		return
	}
	defer listener.Close()

	fmt.Println("[Anubis] internal RPC server is listening on port 8080...")
	rpc.Accept(listener) 
}


func HostConfig(configPath string) {
	cfg, err := loadConfig(configPath)

	if err != nil {
		log.Fatalf("[Anubis Error] error reading config file %s", err)
	}
	databse := selecteDatabase(cfg)
	if databse == "" {
		log.Fatalln("[Anubis Error] database config not found")
	}


	//selecting database from configuration 
	switch databse {
		case "postgres":
			pgxp , err := ConnectPostgres(cfg)
			
			if err != nil {
				log.Fatalf("[Anibis Error] Unable to connect to postgres %s", err)
			}

			pgDatabase := NewPostgresDb(pgxp)
			repository := repository.NewRepository(pgDatabase)
			
			go startRpc(repository)
			selectQueue(cfg, repository)
			
		default:
			log.Fatalln("[Anubis Error] The selected database is not supported")
	}
}







