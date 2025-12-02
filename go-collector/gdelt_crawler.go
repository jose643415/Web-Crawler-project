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


type GDELTResponse struct {
	Articles []GDELTArticle `json:"articles"`
}

type GDELTArticle struct {
	URL           string `json:"url"`
	URLMobile     string `json:"urlmobile"`
	Title         string `json:"title"`
	SeenDate      string `json:"seendate"`
	SocialImg     string `json:"socialimage"`
	Domain        string `json:"domain"`
	Language      string `json:"language"`
	SourceCountry string `json:"sourcecountry"`
}

type GDELTCrawler struct {
	BaseURL string
	Client  *http.Client
}

type KeyValue struct {
	Key   string
	Value int
}


func NewGDELTCrawler() *GDELTCrawler {
	return &GDELTCrawler{
		BaseURL: "https://api.gdeltproject.org/api/v2/doc/doc",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}


// BuscarArticulosMultiLang realiza una búsqueda en GDELT, permitiendo múltiples idiomas.
func (g *GDELTCrawler) BuscarArticulosMultiLang(queryRaw string, idiomas []string, fechaInicio, fechaFin string, maxRecords int) (*GDELTResponse, error) {

	// 1. Construir el filtro de idiomas: (sourceLang:spanish OR sourceLang:english)
	langFilters := make([]string, len(idiomas))
	for i, lang := range idiomas {
		langFilters[i] = fmt.Sprintf("sourceLang:%s", lang)
	}

	langSegment := strings.Join(langFilters, " OR ")

	finalQuery := fmt.Sprintf(`(%s) AND (%s)`, queryRaw, langSegment)

	// 3. Construir URL con parámetros
	params := url.Values{}
	params.Add("query", finalQuery)
	params.Add("mode", "artlist")
	params.Add("maxrecords", fmt.Sprintf("%d", maxRecords))
	params.Add("format", "json")
	params.Add("startdatetime", fechaInicio)
	params.Add("enddatetime", fechaFin)

	fullURL := fmt.Sprintf("%s?%s", g.BaseURL, params.Encode())

	fmt.Printf("Consultando GDELT...\nQuery: %s\nRango: %s - %s\n", finalQuery, fechaInicio, fechaFin)

	// 4. Crear request con User-Agent
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "EthicalCrawler/1.0 (StudentResearch)")

	// 5. Realizar petición
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	// 6. Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error HTTP: status code %d, body: %s", resp.StatusCode, string(body))
	}

	// 7. Parsear JSON (con debug de errores)
	var gdeltResp GDELTResponse
	if err := json.Unmarshal(body, &gdeltResp); err != nil {
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("error parseando JSON: %w. \nRespuesta recibida (Inicio):\n%s", err, preview)
	}

	return &gdeltResp, nil
}


func (g *GDELTCrawler) ExplorarDatos(response *GDELTResponse) {
    // ... (El código de ExplorarDatos es el mismo)
	if response == nil || len(response.Articles) == 0 {
		fmt.Println("\n--- EXPLORACIÓN DE DATOS ---")
		fmt.Println("No se encontraron artículos que coincidan con la búsqueda y los filtros.")
		return
	}

	fmt.Println("\n--- EXPLORACIÓN DE DATOS - GDELT ---")
	fmt.Printf("Total de artículos: %d\n\n", len(response.Articles))

	// Contadores
	dominios := make(map[string]int)
	idiomas := make(map[string]int)
	paises := make(map[string]int)

	for _, art := range response.Articles {
		dominios[art.Domain]++
		idiomas[art.Language]++
		paises[art.SourceCountry]++
	}

	// Mostrar top 10 dominios
	fmt.Println("Top 10 Dominios:")
	topDominios := getTopN(dominios, 10)
	for i, item := range topDominios {
		fmt.Printf("  %2d. %-30s (%d artículos)\n", i+1, item.Key, item.Value)
	}

	// Mostrar idiomas
	fmt.Println("\nDistribución por Idioma:")
	for idioma, count := range idiomas {
		fmt.Printf("  %s: %d\n", idioma, count)
	}

	// Mostrar primeros 5 artículos
	fmt.Println("\nPrimeros 5 Artículos de Muestra:")
	for i, art := range response.Articles {
		if i >= 5 {
			break
		}
		fmt.Printf("\n  %d. Título: %s\n", i+1, art.Title)
		fmt.Printf("      Dominio: %s | Idioma: %s | País: %s\n", art.Domain, art.Language, art.SourceCountry)
		fmt.Printf("      Fecha: %s\n", art.SeenDate)
		fmt.Printf("      URL: %s\n", art.URL)
	}
}

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
	crawler := NewGDELTCrawler()


	query := `"Universidad de Antioquia" OR UdeA` 
    
    // Idiomas y fechas
    idiomasBuscados := []string{"spanish", "english"} 
	fechaInicio := "20230101000000" 
	fechaFin := "20231231235959"    
	maxRecords := 250 

	// Buscar artículos
	response, err := crawler.BuscarArticulosMultiLang(query, idiomasBuscados, fechaInicio, fechaFin, maxRecords)
	if err != nil {
		fmt.Printf("\n--- [ERROR FATAL] ---\n")
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Explorar datos recolectados
	crawler.ExplorarDatos(response)

	fmt.Println("\nExploración completada.")
}