package main
import (
	"fmt"
	"context"
	"github.com/hpharis258/orders-api/application"
	"github.com/redis/go-redis/v9"
)

func main(){
	app := application.New()
	if err := app.Start(context.TODO()); 
	err != nil {
		fmt.Println("Error starting application:", err)
	}
}