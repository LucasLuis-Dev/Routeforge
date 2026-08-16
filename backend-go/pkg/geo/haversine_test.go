package geo

import (
	"testing"
)

func TestCalculateHaversineDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64
	}{
		{
			name:     "Mesma coordenada (distância zero)",
			lat1:     -23.550520,
			lon1:     -46.633308,
			lat2:     -23.550520,
			lon2:     -46.633308,
			expected: 0.0,
		},
		{
			name:     "São Paulo para Rio de Janeiro (~360 km)",
			lat1:     -23.550520, // Praça da Sé, SP
			lon1:     -46.633308,
			lat2:     -22.906847, // Centro, RJ
			lon2:     -43.172896,
			expected: 360.75,
		},
		{
			name:     "Distância curta urbana (~2.5 km)",
			lat1:     -23.561684,
			lon1:     -46.655981,
			lat2:     -23.547778,
			lon2:     -46.635833,
			expected: 2.58,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateHaversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := got - tt.expected
			if diff < 0 {
				diff = -diff
			}
			// Tolerância de 1.0 km devido às variações geodésicas
			if diff > 1.0 {
				t.Errorf("CalculateHaversineDistance() = %v, esperado %v (diferença %v)", got, tt.expected, diff)
			}
		})
	}
}
