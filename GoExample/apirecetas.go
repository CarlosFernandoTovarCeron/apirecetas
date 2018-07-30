
package main

import (
	"log"
	"context"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"

	"firebase.google.com/go"
	"google.golang.org/api/option"
	"cloud.google.com/go/firestore"

	"encoding/json"
	"google.golang.org/api/iterator"

	"os"

	 //"reflect" log.Print(reflect.TypeOf())
)

var client *firestore.Client

type Recipe struct {
    Name string
    Instructions string
}

// Mensaje para el put y el delete
type Msj struct {
	Id string
	Data Recipe
}

func AllRecipesEndPoint(w http.ResponseWriter, r *http.Request) {
	var recipes []map[string]interface {}
	iter := client.Collection("recetas").Documents(context.Background())
	for {
	        doc, err := iter.Next()
	        if err == iterator.Done {
	                break
	        }
	        if err != nil {
	                return
	        }
	        var aux map[string]interface {}
	        aux = doc.Data()
	        aux["id"] = doc.Ref.ID
	        recipes = append(recipes, aux)
	        
	}

	//log.Print(recipes)

	//Retornamos las recetas en json
	js, err := json.Marshal(recipes)

	if err != nil {
  		log.Fatalln(err)
  	}
  	//log.Print(js)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func FindRecipeEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dsnap, err := client.Collection("recetas").Doc(params["id"]).Get(context.Background())
	//log.Print(dsnap.Data())
	if err != nil {
	  	http.Error(w, err.Error(), http.StatusInternalServerError)
	  	return
	}else{

		//Retornamos la receta en json
		js, err := json.Marshal(dsnap.Data())

		if err != nil {
	  		log.Fatalln(err)
	  	}
	  	//log.Print(js)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

}

func CreateRecipeEndPoint(w http.ResponseWriter, r *http.Request) {
	//log.Print(r)
	defer r.Body.Close()
	var recipe Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
	   http.Error(w, err.Error(), http.StatusInternalServerError)
	   return
	}
	//log.Print(recipe)
	_, _, err := client.Collection("recetas").Add(context.Background(), recipe)

	if err != nil {
	   http.Error(w, err.Error(), http.StatusInternalServerError)
	   return
	}

	// Respondemos que se agrego exitosamente
	w.Write([]byte("Receta agregada exitosamente"))
	return
}


func UpdateRecipeEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var msj Msj
	if err := json.NewDecoder(r.Body).Decode(&msj); err != nil {
	   log.Fatalln(err)
	   http.Error(w, err.Error(), http.StatusInternalServerError)
	   return
	}

	_, err := client.Collection("recetas").Doc(msj.Id).Set(context.Background(), msj.Data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}

	// Respondemos que se agrego exitosamente
	w.Write([]byte("Actualizado exitosamente"))
	return
}

func DeleteRecipeEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var msj Msj
	if err := json.NewDecoder(r.Body).Decode(&msj); err != nil {
	   http.Error(w, err.Error(), http.StatusInternalServerError)
	   return
	}

	_, err := client.Collection("recetas").Doc(msj.Id).Delete(context.Background())
	if err != nil {
	   http.Error(w, err.Error(), http.StatusInternalServerError)
	   return
	}else{
		// Respondemos que se elimino exitosamente
		w.Write([]byte("Receta eliminada exitosamente"))
	}

}


func desconectarClienteFirebase(){
	defer client.Close()
}

func main() {
	//Establecemos la conexion
	opt := option.WithCredentialsFile("myfirstpwa-ec1a6-firebase-adminsdk-siciw-ec2397fb09.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
	  log.Fatalln(err)
	}

	//creamos el cliente
	client, err = app.Firestore(context.Background())
	if err != nil {
	  log.Fatalln(err)
	}	

	// imprimimos una receta de prueba (buscar la forma de quitar esto)
	result, err := client.Collection("recetas").Doc("ejemplo").Get(context.Background())
	log.Print(result.Data())

	r := mux.NewRouter()
	r.HandleFunc("/recetas", AllRecipesEndPoint).Methods("GET")
	r.HandleFunc("/recetas", CreateRecipeEndPoint).Methods("POST")
	r.HandleFunc("/recetas", UpdateRecipeEndPoint).Methods("PUT")
	r.HandleFunc("/recetas", DeleteRecipeEndPoint).Methods("DELETE")
	r.HandleFunc("/recetas/{id}", FindRecipeEndpoint).Methods("GET")

	//log.Fatal(http.ListenAndServe(":8000", 
    log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), 
    handlers.LoggingHandler(os.Stdout, handlers.CORS(
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        handlers.AllowedOrigins([]string{"*"}),
        handlers.AllowedHeaders([]string{"Origin", "X-Requested-With", "Content-Type", "Accept"}))(r))))

}