package products

type Product struct {
	Name     string
	Price    float64
	Image    string
	Category string
}

var AllProducts = []Product{
	{Name: "Purificador IBBL Mio Branco", Price: 699.99, Image: "/static/images/purificador.jpg", Category: "purificadores"},
	{Name: "Purificador IBBL FR600", Price: 899.99, Image: "/static/images/purificador.jpg", Category: "purificadores"},
	{Name: "Bebedouro IBBL Compact", Price: 499.99, Image: "/static/images/bebedouro.jpg", Category: "bebedouros"},
	{Name: "Bebedouro Esmaltec EGC35", Price: 599.99, Image: "/static/images/bebedouro.jpg", Category: "bebedouros"},
	{Name: "Válvula Redutora de Pressão 1/4", Price: 45.99, Image: "/static/images/peca.jpg", Category: "pecas"},
	{Name: "Torneira para Purificador", Price: 35.99, Image: "/static/images/peca.jpg", Category: "pecas"},
	{Name: "Refil Gioviale Rpc-01 Lorenzetti", Price: 89.99, Image: "/static/images/refil.jpg", Category: "refis"},
	{Name: "Refil IBBL C+3", Price: 95.99, Image: "/static/images/refil.jpg", Category: "refis"},
}

func GetProductsByCategory(category string) []Product {
	if category == "" {
		return AllProducts
	}
	
	var filtered []Product
	for _, p := range AllProducts {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
