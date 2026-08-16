package geo

import (
	"math"
)

const earthRadiusKM = 6371.0

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

// CalculateHaversineDistance calcula a distância geodésica entre duas coordenadas em km
func CalculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	rLat1 := degreesToRadians(lat1)
	rLat2 := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadiusKM * c
	// Retorna distância arredondada para 2 casas decimais
	return math.Round(distance*100) / 100
}
