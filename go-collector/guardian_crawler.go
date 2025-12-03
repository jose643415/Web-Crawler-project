package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// GuardianResponse mapea el objeto 'response' de la API
type GuardianResponse struct {
	Response struct {
		Status      string           `json:"status"`
		Total       int              `json:"total"`
		PageSize    int              `json:"pageSize"`
		CurrentPage int              `json:"currentPage"`
		Pages       int              `json:"pages"`
		Results     []GuardianArticle `json:"results"`
	} `json:"response"`
}

// GuardianArticle mapea los campos relevantes
type GuardianArticle struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	SectionName   string    `json:"sectionName"`
	WebTitle      string    `json:"webTitle"`
	WebUrl        string    `json:"webUrl"`
	WebPublicationDate time.Time `json:"webPublicationDate"`
}

// GuardianCrawler encapsula la lógica de conexión
type GuardianCrawler struct {
	BaseURL string
	Client  *http.Client
	APIKey  string
}

// KeyValue es una estructura auxiliar para ordenar mapas
type KeyValue struct {
	Key   string
	Value int
}

func NewGuardianCrawler(apiKey string) *GuardianCrawler {
	return &GuardianCrawler{
		BaseURL: "https://content.guardianapis.com/search",
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
		APIKey: apiKey,
	}
}

// BuscarArticulos realiza una búsqueda en The Guardian API.
// La API usa formato ISO 8601 para fechas.
func (g *GuardianCrawler) BuscarArticulos(queryRaw string, fechaInicio, fechaFin string, pageSize int) (*GuardianResponse, error) {

	// 1. Construir la Query: No necesita el operador AND/OR de idioma, 
	// pero sí la expansión de términos.
	// La query aquí se mantiene simple para el parámetro 'q'.
	finalQuery := strings.ReplaceAll(queryRaw, `OR`, `|`)
	finalQuery = strings.ReplaceAll(finalQuery, `"`, ``)
    
	// 2. Construir URL con parámetros
	params := url.Values{}
	params.Add("api-key", g.APIKey)
	params.Add("q", finalQuery)
	// Solo buscamos artículos (no secciones, tags, etc.)
	params.Add("type", "article") 
	params.Add("page-size", fmt.Sprintf("%d", pageSize))

	// Fechas en formato ISO 8601 (YYYY-MM-DD)
	params.Add("from-date", fechaInicio) 
	params.Add("to-date", fechaFin)     
    
    // Filtro de idioma/sección (Guardian no tiene filtro de idioma nativo como NewsAPI)
    // Sin embargo, podemos filtrar por secciones o tags relacionados con Colombia.

	fullURL := fmt.Sprintf("%s?%s", g.BaseURL, params.Encode())

	fmt.Printf("Consultando The Guardian...\nQuery: %s\nRango: %s a %s\n", 
		finalQuery, fechaInicio, fechaFin)
	
	// 3. Realizar petición (no se requiere User-Agent especial para esta API)
	resp, err := g.Client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	// 4. Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error HTTP: status code %d, body: %s", resp.StatusCode, string(body))
	}

	// 5. Parsear JSON
	var apiResp GuardianResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("error parseando JSON: %w. Respuesta recibida (Inicio):\n%s", err, preview)
	}
    
    // Comprobación de status dentro del cuerpo (específico de The Guardian)
    if apiResp.Response.Status != "ok" {
        return nil, fmt.Errorf("error de Guardian API (Status: %s).", apiResp.Response.Status)
    }

	return &apiResp, nil
}


// ExplorarDatosGuardian muestra estadísticas básicas
func (g *GuardianCrawler) ExplorarDatosGuardian(response *GuardianResponse) {
	respData := response.Response
	if len(respData.Results) == 0 {
		fmt.Println("\n--- EXPLORACIÓN DE DATOS ---")
		fmt.Printf("No se encontraron artículos. Total de resultados reportados: %d\n", respData.Total)
		return
	}

	fmt.Println("\n--- EXPLORACIÓN DE DATOS - THE GUARDIAN ---")
	fmt.Printf("Total de artículos encontrados (en el archivo): %d\n", respData.Total)
	fmt.Printf("Artículos recuperados (página): %d\n\n", len(respData.Results))

	// Contador de secciones
	secciones := make(map[string]int)

	for _, art := range respData.Results {
		secciones[art.SectionName]++
	}

	// Mostrar top 5 secciones
	fmt.Println("Top 5 Secciones:")
	topSecciones := getTopN(secciones, 5)
	for i, item := range topSecciones {
		fmt.Printf("  %2d. %-20s (%d artículos)\n", i+1, item.Key, item.Value)
	}
	
	// Mostrar primeros 5 artículos
	fmt.Println("\nPrimeros 5 Artículos de Muestra:")
	for i, art := range respData.Results {
		if i >= 5 {
			break
		}
		fmt.Printf("\n  %d. Título: %s\n", i+1, art.WebTitle)
		fmt.Printf("      Sección: %s\n", art.SectionName)
		fmt.Printf("      Publicado: %s\n", art.WebPublicationDate.Format("2006-01-02"))
		fmt.Printf("      URL: %s\n", art.WebUrl)
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
	apiKey := "04920bd5-2067-419f-9d88-95f9f52551ed" 
    
	crawler := NewGuardianCrawler(apiKey)

	// 1. QUERY: Usamos el formato "OR" y eliminamos las comillas en main.
	// La API de The Guardian usa "|" como OR. Lo convertimos dentro de la función.
	query := `Universidad de Antioquia OR UdeA` 
    
    // 2. RANGO DE FECHAS: Usamos el formato ISO 8601 YYYY-MM-DD
    // Volvemos al 2023 completo para aprovechar el archivo histórico de The Guardian.
	fechaInicio := "2023-01-01" 
	fechaFin := "2023-12-31"    
    
	pageSize := 50 // Artículos a recuperar por página (máx. 50)

	// Buscar artículos
	response, err := crawler.BuscarArticulos(query, fechaInicio, fechaFin, pageSize)
	if err != nil {
		fmt.Printf("\n--- [ERROR FATAL] ---\n")
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Explorar datos recolectados
	crawler.ExplorarDatosGuardian(response)

	fmt.Println("\nExploración completada.")
}