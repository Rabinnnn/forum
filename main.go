package main
import(
	"fmt"
	"os"
)


func main(){
	if len(os.Args) != 1{
		fmt.Println("invalid number of arguments.")
		fmt.Println("Usage: go run .")
		return
	}

	db, err := InitializeDB()
	if err != nil{
		fmt.Printf("failed to initialize database: %v", err)
	}
	defer db.Close()
}