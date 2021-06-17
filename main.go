package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	//"github.com/rs/cors"
	"github.com/dgrijalva/jwt-go"
)

// Creo key con palabra secreta
var jwtKey = []byte("my_secret_key")

//usuarios hardcodeados
var users = map[string]string{
	"admin":  "12345",
	"sergio": "demodemo",
}

// Estructura para credenciales
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Creo estructura JWT.
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login para generar token
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expectedPassword, ok := users[creds.Username]

	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//token expirará cada 5 minutos
	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Genero cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenString)
}

//estructura basica tipo clave/valor (id/nombre)
type Kv struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Property struct
type Property struct {
	Id               int    `json:"id"`
	Title            string `json:"title"`
	Property_type    Kv     `json:"property_type"`
	Transaction_type Kv     `json:"transaction_type"`
	Currency         Kv     `json:"currency"`
	Address          string `json:"address"`
	Address_number   string `json:"address_number"`
	City             Kv     `json:"city"`
	State            Kv     `json:"state"`
	Country          Kv     `json:"country"`
	Neighborhood     string `json:"neighborhood"`
	Description      string `json:"description"`
	Rooms            int    `json:"rooms"`
	Bedrooms         int    `json:"bedrooms"`
	Bathrooms        int    `json:"bathrooms"`
	Garages          int    `json:"garage"`
	M2               int    `json:"m2"`
	M2_covered       int    `json:"m2_covered"`
	Year             int    `json:"year"`
	Price            int64  `json:"price"`
	Status           string `json:"status"`
}

var properties []Property

func getProperties(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tknStr := c.Value

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(properties)
}

// Todos las propiedades sin jwt
/*func getProperties(w http.ResponseWriter, router *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(properties)
}*/

// Propiedad individual
func getProperty(w http.ResponseWriter, router *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(router)
	// Recorro array de propiedades buscando el que coincida en id
	// en produccion no hay loop, sino que se consulta a DB
	id := 0
	for _, item := range properties {
		id, _ = strconv.Atoi(params["id"])
		if item.Id == id {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Property{})
}

// Nueva Propiedad
func createProperty(w http.ResponseWriter, router *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var property Property
	_ = json.NewDecoder(router.Body).Decode(&property)
	aux := 1
	for _, item := range properties {
		if aux < item.Id {
			aux = item.Id
		}
	}
	property.Id = aux + 1
	properties = append(properties, property)
	json.NewEncoder(w).Encode(property)
}

// Actualizar Propiedad
func updateProperty(w http.ResponseWriter, router *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(router)
	id := 0
	var property Property
	for idx, item := range properties {
		id, _ = strconv.Atoi(params["id"])
		if item.Id == id {
			property = properties[idx]
			properties = append(properties[:idx], properties[idx+1:]...)
			var aux Property
			_ = json.NewDecoder(router.Body).Decode(&aux)
			property.Status = aux.Status
			properties = append(properties, property)
			json.NewEncoder(w).Encode(property)
			return
		}
	}
}

// Borrar Propiedad
func deleteProperty(w http.ResponseWriter, router *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(router)
	id := 0
	for idx, item := range properties {
		id, _ = strconv.Atoi(params["id"])
		if item.Id == id {
			properties = append(properties[:idx], properties[idx+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(properties)
}

func main() {
	// Inicio router
	r := mux.NewRouter()

	// Datos hardcodeados
	properties = append(properties, Property{Id: 1000, Title: "Prueba 1", Property_type: Kv{Id: 8, Name: "Departamento"}, Transaction_type: Kv{Id: 2, Name: "Venta"}, Currency: Kv{Id: 2, Name: "$"}, Address: "Luro", Address_number: "3009", City: Kv{Id: 569943, Name: "Mar del Plata"}, State: Kv{Id: 1818, Name: "Buenos Aires"}, Country: Kv{Id: 5, Name: "Argentina"}, Neighborhood: "Centro", Rooms: 2, Bedrooms: 2, Bathrooms: 1, Garages: 1, M2: 300, M2_covered: 300, Year: 1998, Price: 1000000, Status: "available"})
	properties = append(properties, Property{Id: 1001, Title: "", Property_type: Kv{Id: 8, Name: "Departamento"}, Transaction_type: Kv{Id: 2, Name: "Venta"}, Currency: Kv{Id: 1, Name: "u$s"}, Address: "Salta / Libertad y Maip\u00fa", Address_number: "", City: Kv{Id: 569943, Name: "Mar del Plata"}, State: Kv{Id: 1818, Name: "Buenos Aires"}, Country: Kv{Id: 5, Name: "Argentina"}, Neighborhood: "La Perla", Description: "Departamento original en perfecto estado de conservaci\u00f3n y uso. Cocina separada con lavadero. Muy luminoso. Buen placard en el ambiente, el cual es divisible y est\u00e1 orientado a la calle, al igual que...", Rooms: 1, Bedrooms: 1, Bathrooms: 1, Garages: 1, M2: 33, M2_covered: 28, Year: 2015, Price: 8250000, Status: "available"})
	properties = append(properties, Property{Id: 1002, Title: "Local en La Perla", Property_type: Kv{Id: 16, Name: "Local"}, Transaction_type: Kv{Id: 2, Name: "Venta"}, Currency: Kv{Id: 1, Name: "u$s"}, Address: "Independencia", Address_number: "873", City: Kv{Id: 569943, Name: "Mar del Plata"}, State: Kv{Id: 1818, Name: "Buenos Aires"}, Country: Kv{Id: 5, Name: "Argentina"}, Neighborhood: "La Perla", Description: "Se trata de un local al frente, sobre la avenida Independencia. El mismo est\u00e1 en excelente estado. Tiene unos 43 mts. Cubiertos aproximadamente.\r\nConsta de un amplio ambiente irregular al frente, un m...", Rooms: 2, Bedrooms: 1, Bathrooms: 1, Garages: 1, M2: 65, M2_covered: 43, Year: 2028, Price: 6550000, Status: "available"})
	properties = append(properties, Property{Id: 1003, Title: "Local en La Perla", Property_type: Kv{Id: 16, Name: "Local"}, Transaction_type: Kv{Id: 2, Name: "Venta"}, Currency: Kv{Id: 1, Name: "u$s"}, Address: "Independencia", Address_number: "873", City: Kv{Id: 569943, Name: "Mar del Plata"}, State: Kv{Id: 1818, Name: "Buenos Aires"}, Country: Kv{Id: 5, Name: "Argentina"}, Neighborhood: "La Perla", Description: "Se trata de un local al frente, sobre la avenida Independencia. El mismo est\u00e1 en excelente estado. Tiene unos 43 mts. Cubiertos aproximadamente.\r\nConsta de un amplio ambiente irregular al frente, un m...", Rooms: 1, Bedrooms: 1, Bathrooms: 1, Garages: 0, M2: 43, M2_covered: 43, Year: 1998, Price: 6544000, Status: "available"})

	// Endpoints
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/properties", getProperties).Methods("GET")
	r.HandleFunc("/properties/{id}", getProperty).Methods("GET")
	r.HandleFunc("/properties", createProperty).Methods("POST")
	r.HandleFunc("/properties/{id}", updateProperty).Methods("PUT")
	r.HandleFunc("/properties/{id}", deleteProperty).Methods("DELETE")

	// CORS
	/*corsWrapper := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})*/

	// Server
	fmt.Println("Inicio server: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
