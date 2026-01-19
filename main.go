package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/rmccullagh/weather-api/docs"
	"github.com/rmccullagh/weather-api/models"
	"github.com/rmccullagh/weather-api/services"
	"github.com/rmccullagh/weather-api/utils"
	httpSwagger "github.com/swaggo/http-swagger"
)

// GetForecast
//
//	@Summary		Returns the forecasted weather by latitude and longitude coordinates
//	@Description	Get Forecast By Coordinates
//	@ID				get-forecast-by-coordinates
//	@Produce		json
//	@Param			latitude	 path	    number true	"The latitude of the desired location  (e.g. 39.7456)" Format(float)
//	@Param			longitude	 path	    number true	"The longitude of the desired location  (e.g. -97.0892)" Format(flaot)
//	@Success		200		{object}	models.Forecast
//	@Failure	    500		{object}	models.APIError
//	@Router			/v1/forecasts/{latitude}/{longitude} [get]
func GetForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	latitude := chi.URLParam(r, "latitude")
	longitude := chi.URLParam(r, "longitude")

	client := services.NewClient()
	forecast, err := client.GetForecast(latitude, longitude)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.JSONResponse(w, models.APIError{Message: err.Error()})
		return
	}

	utils.JSONResponse(w, forecast)
}

func RedirectRootToSwagger(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/index.html", http.StatusTemporaryRedirect)
}

func SwaggerHandler() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // The url pointing to API definition
	)
}

// @title Weather API
// @version 1.0
// @description This is an HTTP server that serves the forecasted weather
// @title Weather API
// @version 1.0
// @description This is an HTTP server that serves the forecasted weather

// @host localhost:8080
// @BasePath /
func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Redirect root to swagger docs
	router.Get("/", RedirectRootToSwagger)

	router.Route("/v1", func(r chi.Router) {
		r.Get("/forecasts/{latitude}/{longitude}", GetForecast)
	})

	router.Get("/swagger/*", SwaggerHandler())

	log.Println("Go to http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("unable to listen on port 8080")
	}
}
