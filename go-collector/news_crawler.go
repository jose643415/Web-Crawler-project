package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

// NewsAPIResponse mapea la respuesta principal de NewsAPI
type NewsAPIResponse struct {
	Status       string          `json:"status"`
	TotalResults int             `json:"totalResults"`
	Articles     []NewsAPIArticle `json:"articles"`
}

// NewsAPIArticle mapea los campos relevantes de cada artículo
type NewsAPIArticle struct {
	Source struct {
		Name string `json:"name"`
	} `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// NewsAPICrawler encapsula la lógica de conexión
type NewsAPICrawler struct {
	BaseURL string
	Client  *http.Client
	APIKey  string
}

// KeyValue es una estructura auxiliar para ordenar mapas (misma que GDELT)
type KeyValue struct {
	Key   string
	Value int
}


func NewNewsAPICrawler(apiKey string) *NewsAPICrawler {
	return &NewsAPICrawler{
		BaseURL: "https://newsapi.org/v2/everything",
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
		APIKey: apiKey,
	}
}


// BuscarArticulos realiza una búsqueda en NewsAPI.
// NewsAPI no usa "sourceLang", sino el parámetro "language" con códigos ISO 639-1 de dos letras.
// Los idiomas se pasan como una cadena de dos letras separadas por comas (ej: "es,en").
func (n *NewsAPICrawler) BuscarArticulos(queryRaw, idiomasCSV, fechaInicio, fechaFin string, pageSize int) (*NewsAPIResponse, error) {

	// 1. Construir la Query: NewsAPI soporta operadores AND/OR.
	// La query debe ser simple sin la sintaxis especial de GDELT.
	finalQuery := fmt.Sprintf(`"Universidad de Antioquia" OR UdeA`)

	// 2. Construir URL con parámetros
	params := url.Values{}
	params.Add("q", finalQuery)
	params.Add("language", idiomasCSV) // "es,en"
	params.Add("sortBy", "publishedAt")
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))
	
	// Fechas deben estar en formato ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)
	// Asumimos que fechaInicio y fechaFin ya vienen en ese formato o similar
	// Si vienen en formato GDELT (YYYYMMDDHHMMSS), esto fallará, por eso lo ajustamos en main.
	params.Add("from", fechaInicio)
	params.Add("to", fechaFin)

	fullURL := fmt.Sprintf("%s?%s", n.BaseURL, params.Encode())

	fmt.Printf("Consultando NewsAPI...\nQuery: %s\nIdiomas: %s\nRango: %s a %s\n", 
		finalQuery, idiomasCSV, fechaInicio, fechaFin)

	// 3. Crear request con API Key en el Header (es la forma preferida)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	// Agregar la API Key y User-Agent
	req.Header.Set("X-Api-Key", n.APIKey)
	req.Header.Set("User-Agent", "EthicalCrawlerNews/1.0")

	// 4. Realizar petición
	resp, err := n.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	// 5. Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	// 6. Parsear JSON
	var apiResp NewsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("error parseando JSON: %w. \nRespuesta recibida (Inicio):\n%s", err, preview)
	}
    
    // NewsAPI devuelve el status en el cuerpo, no solo en el HTTP status code
    if apiResp.Status != "ok" {
        // En caso de error de API (ej: API Key inválida, límite de fechas)
        return nil, fmt.Errorf("error de NewsAPI (Status: %s). El cuerpo puede tener más detalles: %s", apiResp.Status, string(body))
    }


	return &apiResp, nil
}


// ExplorarDatosNewsAPI muestra estadísticas básicas
func (n *NewsAPICrawler) ExplorarDatosNewsAPI(response *NewsAPIResponse) {
	if response == nil || len(response.Articles) == 0 {
		fmt.Println("\n--- EXPLORACIÓN DE DATOS ---")
		fmt.Println("No se encontraron artículos que coincidan con la búsqueda.")
		return
	}

	fmt.Println("\n--- EXPLORACIÓN DE DATOS - NEWSAPI ---")
	fmt.Printf("Total de artículos encontrados: %d\n\n", response.TotalResults)
	fmt.Printf("Artículos recuperados (página): %d\n\n", len(response.Articles))

	// Contador de fuentes (Sources)
	fuentes := make(map[string]int)

	for _, art := range response.Articles {
		fuentes[art.Source.Name]++
	}

	// Mostrar top 10 fuentes
	fmt.Println("Top 10 Fuentes:")
	topFuentes := getTopN(fuentes, 10)
	for i, item := range topFuentes {
		fmt.Printf("  %2d. %-30s (%d artículos)\n", i+1, item.Key, item.Value)
	}
	
	// Mostrar primeros 5 artículos
	fmt.Println("\nPrimeros 5 Artículos de Muestra:")
	for i, art := range response.Articles {
		if i >= 5 {
			break
		}
		fmt.Printf("\n  %d. Título: %s\n", i+1, art.Title)
		fmt.Printf("      Fuente: %s | Autor: %s\n", art.Source.Name, art.Author)
		fmt.Printf("      Publicado: %s\n", art.PublishedAt.Format("2006-01-02 15:04"))
		fmt.Printf("      URL: %s\n", art.URL)
	}
}

// getTopN (misma función auxiliar)
func getTopN(m map[string]int, n int) []KeyValue {
	var kvList []KeyValue
	for k, v := range m {
		kvList = append(kvList, KeyValue{k, v})
	}

	sort.Slice(kvList, func(i, j int) bool {
		return kvList[i].Value > kvList[j].Value
	})

	if n > len(kvList) {
		n = len(kvList)
	}
	return kvList[:n]
}


func main() {
	apiKey := "92437566c60d4a14b89ca3c20960b8ed" 
    
	crawler := NewNewsAPICrawler(apiKey)

	query := `"Universidad de Antioquia" OR UdeA` 
    
	idiomasCSV := "es,en" 
    
    now := time.Now() 
    
	fechaInicio := now.AddDate(0, 0, -30).Format("2006-01-02T15:04:05") 
	fechaFin := now.Format("2006-01-02T15:04:05")    
    
	pageSize := 50 

	response, err := crawler.BuscarArticulos(query, idiomasCSV, fechaInicio, fechaFin, pageSize)
	if err != nil {
		fmt.Printf("\n--- [ERROR FATAL] ---\n")
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Explorar datos recolectados
	crawler.ExplorarDatosNewsAPI(response)

	fmt.Println("\nExploración completada.")
}