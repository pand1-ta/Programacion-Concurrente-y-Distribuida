package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Struct para almacenar los datos que vamos a procesar
type UserReview struct {
	Recommended     bool
	NumReviews      int
	PlaytimeForever int
}

// Función para leer el archivo CSV y devolver un slice de UserReview
func LeerCSV(nombreArchivo string) ([]UserReview, error) {
	// Abrimos el archivo
	archivo, err := os.Open(nombreArchivo)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer archivo.Close()

	lector := csv.NewReader(archivo)

	// Leemos el header
	_, err = lector.Read()
	if err != nil {
		return nil, fmt.Errorf("error al leer el encabezado: %v", err)
	}

	var reviews []UserReview

	for {
		registro, err := lector.Read()
		if err != nil {
			break // EOF
		}

		// Ajusta los índices según el orden de tu archivo CSV
		recommendedStr := registro[8] // recommended
		numReviewsStr := registro[18] // author.num_reviews
		playtimeStr := registro[19]   // author.playtime_forever

		// Parseo de recommended (bool)
		recommended := recommendedStr == "True"

		// Parseo de numReviews (int)
		numReviews, err := strconv.Atoi(numReviewsStr)
		if err != nil {
			numReviews = 0
		}

		// Parseo de playtime (float a int)
		playtimeFloat, err := strconv.ParseFloat(playtimeStr, 64)
		playtime := 0
		if err == nil {
			playtime = int(playtimeFloat)
		}

		// Creamos el struct
		review := UserReview{
			Recommended:     recommended,
			NumReviews:      numReviews,
			PlaytimeForever: playtime,
		}

		reviews = append(reviews, review)
	}

	return reviews, nil
}

// Función para contar las recomendaciones positivas y negativas
func ContarRecomendaciones(data []UserReview) (int, int) {
	var totalTrue, totalFalse int

	for _, review := range data {
		if review.Recommended {
			totalTrue++
		} else {
			totalFalse++
		}
	}

	return totalTrue, totalFalse
}

// Función para calcular el promedio de num_reviews
func PromedioNumReviews(data []UserReview) float64 {
	if len(data) == 0 {
		return 0
	}

	var suma int

	for _, review := range data {
		suma += review.NumReviews
	}

	promedio := float64(suma) / float64(len(data))
	return promedio
}

// Función para calcular el promedio de playtime_forever
func PromedioPlaytime(data []UserReview) float64 {
	if len(data) == 0 {
		return 0
	}

	var suma int

	for _, review := range data {
		suma += review.PlaytimeForever
	}

	promedio := float64(suma) / float64(len(data))
	return promedio
}

// Función para agrupar usuarios por rangos de playtime_forever
func AgruparPorRangosDePlaytime(data []UserReview) map[string]int {
	// Crear un mapa para los rangos
	rangos := map[string]int{
		"0-100":    0,
		"101-500":  0,
		"501-1000": 0,
		"1001+":    0,
	}

	// Clasificar usuarios en rangos
	for _, review := range data {
		pt := review.PlaytimeForever

		switch {
		case pt <= 100:
			rangos["0-100"]++
		case pt <= 500:
			rangos["101-500"]++
		case pt <= 1000:
			rangos["501-1000"]++
		default:
			rangos["1001+"]++
		}
	}

	return rangos
}

// Función para encontrar usuarios que tienen más de 1000 horas de juego pero no recomiendan el juego
func UsuariosConMuchoJuegoPeroNoRecomiendan(data []UserReview) []UserReview {
	var casos []UserReview

	for _, review := range data {
		if !review.Recommended && review.PlaytimeForever > 1000 {
			casos = append(casos, review)
		}
	}

	return casos
}

// Función para calcular el promedio de playtime_forever para usuarios que recomiendan y no recomiendan
func PromedioPlaytimePorRecomendacion(data []UserReview) (float64, float64) {
	var sumaSi, sumaNo int
	var totalSi, totalNo int

	for _, review := range data {
		if review.Recommended {
			sumaSi += review.PlaytimeForever
			totalSi++
		} else {
			sumaNo += review.PlaytimeForever
			totalNo++
		}
	}

	var promSi, promNo float64

	if totalSi > 0 {
		promSi = float64(sumaSi) / float64(totalSi)
	}
	if totalNo > 0 {
		promNo = float64(sumaNo) / float64(totalNo)
	}

	return promSi, promNo
}

// Función para exportar el resumen a un archivo de texto
func ExportarResumen(nombreArchivo string, total int, recomiendan int, noRecomiendan int,
	promReviews float64, promPlaytime float64,
	rangos map[string]int, outliers int,
	promSi float64, promNo float64) error {

	contenido := fmt.Sprintf(`📄 RESUMEN DEL ANÁLISIS DE STEAM REVIEWS

Total de usuarios procesados: %d

Recomendaciones:
- Recomiendan: %d (%.2f%%)
- No recomiendan: %d (%.2f%%)

Promedio de reseñas por usuario: %.2f
Promedio de tiempo jugado: %.2f minutos

Promedio de tiempo jugado según recomendación:
- Recomiendan: %.2f minutos
- No recomiendan: %.2f minutos

Distribución por rangos de tiempo jugado:
`, total, recomiendan, float64(recomiendan)*100/float64(total),
		noRecomiendan, float64(noRecomiendan)*100/float64(total),
		promReviews, promPlaytime,
		promSi, promNo)

	for rango, cantidad := range rangos {
		contenido += fmt.Sprintf("- %s minutos: %d usuarios\n", rango, cantidad)
	}

	contenido += fmt.Sprintf("\nUsuarios que jugaron más de 1000 min y no recomiendan: %d\n", outliers)

	return os.WriteFile(nombreArchivo, []byte(contenido), 0644)
}

// Función principal
func main() {
	nombreArchivo := "EDA/steam_reviews.csv"

	// Paso 1: Leer el archivo CSV
	reviews, err := LeerCSV(nombreArchivo)
	if err != nil {
		fmt.Println("Error al leer el archivo:", err)
		return
	}

	fmt.Println("Dataset cargado correctamente.")
	fmt.Printf("Total de usuarios procesados: %d\n\n", len(reviews))

	// Paso 2: Contar recomendaciones
	recomiendan, noRecomiendan := ContarRecomendaciones(reviews)
	fmt.Println("Recomendaciones:")
	fmt.Printf("Recomiendan: %d (%.2f%%)\n", recomiendan, float64(recomiendan)*100/float64(len(reviews)))
	fmt.Printf("No recomiendan: %d (%.2f%%)\n\n", noRecomiendan, float64(noRecomiendan)*100/float64(len(reviews)))

	// Paso 3: Promedio de reseñas por usuario
	promReviews := PromedioNumReviews(reviews)
	fmt.Printf("Promedio de reseñas por usuario: %.2f\n", promReviews)

	// Paso 4: Promedio de tiempo jugado
	promPlaytime := PromedioPlaytime(reviews)
	fmt.Printf("Promedio de tiempo jugado: %.2f minutos\n\n", promPlaytime)

	// Paso 5: Agrupar por rangos de juego
	rangos := AgruparPorRangosDePlaytime(reviews)
	fmt.Println("Distribución por rangos de tiempo jugado:")
	for rango, cantidad := range rangos {
		fmt.Printf("- %s minutos: %d usuarios\n", rango, cantidad)
	}
	fmt.Println()

	// Paso 6: Usuarios con mucho juego y no recomiendan
	atipicos := UsuariosConMuchoJuegoPeroNoRecomiendan(reviews)
	fmt.Printf("Usuarios que jugaron más de 1000 min y no recomiendan: %d\n", len(atipicos))

	// Paso 7: Promedio de playtime por recomendación
	promSi, promNo := PromedioPlaytimePorRecomendacion(reviews)
	fmt.Printf("Promedio de tiempo jugado para quienes recomiendan: %.2f minutos\n", promSi)
	fmt.Printf("Promedio de tiempo jugado para quienes no recomiendan: %.2f minutos\n", promNo)

	// Paso 8: Exportar resumen a archivo de texto
	timestamp := time.Now().Format("20060102_150405")
	nombreResumen := fmt.Sprintf("EDA/resumen_steam_%s.txt", timestamp)
	err = ExportarResumen(nombreResumen, len(reviews), recomiendan, noRecomiendan,
		promReviews, promPlaytime,
		rangos, len(atipicos),
		promSi, promNo)
	if err != nil {
		fmt.Println("Error al exportar el resumen:", err)
		return
	}
	fmt.Printf("\nResumen exportado a %s\n", nombreResumen)

}
