package main
import (
	"fmt"
	"github.com/hpharis258/GoOrdersAPI/application"
)

func main(){
	app := application.New()
	if err := app.Start(context.TODO()); 
	err != nil {
		fmt.Println("Error starting application:", err)
	}
}