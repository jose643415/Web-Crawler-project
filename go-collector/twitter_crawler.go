package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)


type XResponse struct {
	Data []Tweet `json:"data"`
	Meta XMeta   `json:"meta"`
}

// Estructura para capturar las métricas de interacción
type PublicMetrics struct {
	RetweetCount int `json:"retweet_count"`
	LikeCount    int `json:"like_count"`
	ReplyCount   int `json:"reply_count"`
	QuoteCount   int `json:"quote_count"`
}

// Tweet actualizado para incluir las métricas
type Tweet struct {
	ID            string        `json:"id"`
	Text          string        `json:"text"`
	CreatedAt     time.Time     `json:"created_at"`
	PublicMetrics PublicMetrics `json:"public_metrics"` 
}

type XMeta struct {
	NewestID    string `json:"newest_id"`
	OldestID    string `json:"oldest_id"`
	ResultCount int    `json:"result_count"`
	NextToken   string `json:"next_token"`
}

type XCrawler struct {
	BaseURL     string
	Client      *http.Client
	BearerToken string
}

type KeyValue struct {
	Key   string
	Value int
}


func NewXCrawler(bearerToken string) *XCrawler {
	return &XCrawler{
		BaseURL: "https://api.twitter.com/2/tweets/search/recent",
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
		BearerToken: bearerToken,
	}
}


func (x *XCrawler) BuscarTweets(queryRaw string, maxResults int, startTime, endTime string) (*XResponse, error) {

	// Query: ("Universidad de Antioquia" OR UdeA) investigación lang:es -is:retweet
	finalQuery := fmt.Sprintf(`(%s) investigación lang:es -is:retweet`, queryRaw)

	// 1. Construir URL con parámetros
	params := url.Values{}
	params.Add("query", finalQuery)
	// 🎯 ACTUALIZACIÓN: Incluir 'public_metrics' para obtener el conteo de retweets
	params.Add("tweet.fields", "created_at,public_metrics") 
	params.Add("max_results", fmt.Sprintf("%d", maxResults)) 
    
	// 🎯 ACTUALIZACIÓN: Añadir parámetros de tiempo
	params.Add("start_time", startTime) 
	params.Add("end_time", endTime)   

	fullURL := fmt.Sprintf("%s?%s", x.BaseURL, params.Encode())

	fmt.Printf("Consultando X (Reciente)...\nQuery: %s\nRango: %s a %s\n", finalQuery, startTime, endTime)

	// 2. Crear request y añadir Bearer Token
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+x.BearerToken)
	req.Header.Set("User-Agent", "EthicalXCrawler/1.0 (StudentResearch)")

	// 3. Realizar petición y manejo de errores
	resp, err := x.Client.Do(req)
	// ... (Resto del código de petición y manejo de errores se omite por brevedad) ...
    
	if err != nil {
		return nil, fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error HTTP: status code %d. Respuesta de X:\n%s", resp.StatusCode, string(body))
	}

	var apiResp XResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("error parseando JSON: %w. Respuesta recibida:\n%s", err, preview)
	}

	return &apiResp, nil
}

func ExplorarDatosX(response *XResponse) {
	if response == nil || response.Meta.ResultCount == 0 {
		fmt.Println("\n--- EXPLORACIÓN DE DATOS X ---")
		fmt.Println("No se encontraron tweets que coincidan con la búsqueda.")
		return
	}

	fmt.Println("\n--- EXPLORACIÓN DE DATOS - X (Últimos 7 Días) ---")
	fmt.Printf("Total de tweets encontrados: %d\n", response.Meta.ResultCount)
	fmt.Printf("Tweets recuperados: %d\n\n", len(response.Data))

	// Mostrar los primeros 5 tweets
	fmt.Println("Primeros 5 Tweets de Muestra:")
	for i, tweet := range response.Data {
		if i >= 5 {
			break
		}
		fmt.Printf("\n  %d. ID: %s\n", i+1, tweet.ID)
		fmt.Printf("      Fecha: %s\n", tweet.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("      Compartidos/Retweets: %d\n", tweet.PublicMetrics.RetweetCount) // 🎯 NUEVO DATO
		fmt.Printf("      Likes: %d | Respuestas: %d\n", tweet.PublicMetrics.LikeCount, tweet.PublicMetrics.ReplyCount)
		fmt.Printf("      Texto: %s\n", tweet.Text)
	}
}


func main() {

	bearerToken := "AAAAAAAAAAAAAAAAAAAAAJW%2F5gEAAAAAr4HJjlMgtehsrwTzfC1IfxsiVmw%3DkSvbyiSiONuREZivSiIHh3x1VsyMXZh6iUgHCxl8uy6URvOPe7" 
    
	crawler := NewXCrawler(bearerToken)

	query := `"Universidad de Antioquia" OR UdeA` 
       
    now := time.Now().UTC().Add(-1 * time.Minute) 
    
    sevenDaysAgo := now.AddDate(0, 0, -7) 
    
    // Formato ISO 8601 con 'Z' para UTC
    startTime := sevenDaysAgo.Format("2006-01-02T15:04:05Z") 
    endTime := now.Format("2006-01-02T15:04:05Z")    
    
	maxResults := 50

	// Buscar tweets
	response, err := crawler.BuscarTweets(query, maxResults, startTime, endTime)
	if err != nil {
		fmt.Printf("\n--- [ERROR FATAL X] ---\n")
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Explorar datos recolectados
	ExplorarDatosX(response)

	fmt.Println("\nExploración completada.")
}