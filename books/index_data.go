package books

import "html/template"

type IndexData struct {
	Data map[string]template.HTML
}

var (
	LangData = map[string]IndexData{
		"en": indexEn,
		"es": indexEs,
	}
	indexEn = IndexData{
		Data: map[string]template.HTML{
			"PageTitle":       "Library &mdash; API",
			"SubTitle":        "This is the api for the Library app located at ",
			"OtherLang":       "Versión en español.",
			"OtherLangLink":   "/es",
			"AvailableRoutes": "Available routes",
			"Name":            "Name",
			"Route":           "Route",
			"Example":         "Example",
			"Notes":           "Notes",
			"BookList":        "Book list",
			"BookByID":        "Book by id",
			"BookSearch":      "Book search by title or author",
			"qRequired":       "The <code>q</code> parameter is required",
			"PageList":        "Page list",
			"PageByID":        "Page by id",
			"PageParam":       "Page param search",
			"paramsRequired":  "Both the <code>book-id</code> and the <code>page-number</code> query params are required.",
		},
	}
	indexEs = IndexData{
		Data: map[string]template.HTML{
			"PageTitle":       "Biblioteca &mdash; API",
			"SubTitle":        "Este es el API para la aplicación de biblioteca localizada en ",
			"OtherLang":       "English version.",
			"OtherLangLink":   "/en",
			"AvailableRoutes": "Rutas disponibles:",
			"Name":            "Nombre",
			"Route":           "Ruta",
			"Example":         "Ejemplo",
			"Notes":           "Notas",
			"BookList":        "Listado de libros",
			"BookByID":        "Libros por id",
			"BookSearch":      "Búsqueda de libros por título o autor",
			"qRequired":       "El parámetro <code>q</code> es requerido.",
			"PageList":        "Listado de páginas",
			"PageByID":        "Páginas por id",
			"PageParam":       "Búsqueda de páginas por parámetros",
			"paramsRequired":  "Ambos parámetros, <code>book-id</code> y <code>page-number</code> son requeridos.",
		},
	}
)
