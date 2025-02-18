package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// Structure pour stocker les données des pilotes depuis l'API Ergast
type Driver struct {
	DriverID    string `json:"driverId"`
	Code        string `json:"code"`
	FirstName   string `json:"givenName"`
	LastName    string `json:"familyName"`
	Nationality string `json:"nationality"`
	ImageURL    string // Image du pilote (ajouté manuellement)
}

// Structure pour parser la réponse JSON de l'API Ergast
type ErgastResponse struct {
	MRData struct {
		DriverTable struct {
			Drivers []Driver `json:"Drivers"`
		} `json:"DriverTable"`
	} `json:"MRData"`
}

// Structure pour transmettre les données au template HTML
type TemplateData struct {
	Drivers []Driver
}

var templates *template.Template

// Charger les templates
func loadTemplates() {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("❌ Erreur lors du chargement des templates: %v", err)
	}
	templates = tmpl
	log.Println("✅ Templates chargés avec succès")
}

// Ajoute une image fictive aux pilotes
func fetchDriverImage(id string) string {
	return fmt.Sprintf("https://www.formula1.com/content/dam/fom-website/drivers/%s.jpg", id)
}

// Fonction pour récupérer la liste des pilotes depuis l'API Ergast
func fetchDrivers(year string) ([]Driver, error) {
	url := fmt.Sprintf("https://ergast.com/api/f1/%s/drivers.json", year)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Erreur lors de la récupération des pilotes: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var ergastData ErgastResponse
	if err := json.NewDecoder(resp.Body).Decode(&ergastData); err != nil {
		log.Printf("❌ Erreur de décodage JSON: %v", err)
		return nil, err
	}

	drivers := ergastData.MRData.DriverTable.Drivers

	// Ajouter une image fictive pour chaque pilote
	for i := range drivers {
		drivers[i].ImageURL = fetchDriverImage(drivers[i].DriverID)
	}

	log.Printf("✅ Pilotes récupérés pour l'année %s: %+v", year, drivers)
	return drivers, nil
}

// Handler API pour récupérer les pilotes selon l'année
func driversAPIHandler(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" {
		http.Error(w, "Paramètre 'year' manquant", http.StatusBadRequest)
		return
	}

	drivers, err := fetchDrivers(year)
	if err != nil {
		http.Error(w, "❌ Erreur lors de la récupération des pilotes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

// Handler pour afficher la page des pilotes
func driversPageHandler(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" {
		year = "2024" // Valeur par défaut
	}

	drivers, err := fetchDrivers(year)
	if err != nil {
		http.Error(w, "❌ Erreur lors de la récupération des pilotes", http.StatusInternalServerError)
		return
	}

	data := TemplateData{Drivers: drivers}
	err = templates.ExecuteTemplate(w, "drivers", data)
	if err != nil {
		http.Error(w, "❌ Erreur lors de l'exécution du template: "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Charger les templates
	loadTemplates()

	// Servir les fichiers statiques (CSS, images, etc.)
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/drivers", driversPageHandler)
	http.HandleFunc("/api/drivers", driversAPIHandler)

	// Démarrer le serveur
	fmt.Println("🚀 Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil)) // log.Fatal pour afficher les erreurs
}
