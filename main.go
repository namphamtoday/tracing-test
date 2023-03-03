package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type JaegerConfig struct {
	ServiceName        string `json:"service_name" mapstructure:"service_name" envconfig:"service_name"`
	Environment        string `json:"environment" mapstructure:"environment" envconfig:"env"`
	Id                 int64  `json:"id" mapstructure:"id" envconfig:"id"`
	UrlJaegerCollector string `json:"url_Jaeger_Collector" mapstructure:"url_Jaeger_Collector" envconfig:"url_Jaeger_Collector"`
	System             string `json:"system" mapstructure:"system" envconfig:"system"`
}

func NewJaegerExport(conf *JaegerConfig) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(conf.UrlJaegerCollector)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(conf.ServiceName),
			attribute.String("Environment", conf.Environment),
			attribute.Int64("ID", int64(conf.Id)),
			attribute.String("System", conf.System),
		)),
	)
	return tp, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	var jaegerConf JaegerConfig
	if err := envconfig.Process("", &jaegerConf); err != nil {
		fmt.Printf("Failed to process env var: %v\n", err)
		return
	}

	tp, err := NewJaegerExport(&jaegerConf)
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	// Initialize the Echo router
	ech := echo.New()
	// Add a middleware to instrument all requests with OpenTelemetry
	ech.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Start a new span for the incoming request
			tr := tp.Tracer("example/echo")
			_, span := tr.Start(c.Request().Context(), c.Request().Method+" "+c.Request().URL.Path, trace.WithSpanKind(trace.SpanKindServer))

			// Set some common attributes on the span
			span.SetAttributes(semconv.HTTPMethodKey.String(c.Request().Method))
			span.SetAttributes(semconv.HTTPURLKey.String(c.Request().URL.Path))
			span.SetAttributes(semconv.NetHostIPKey.String(c.Request().RemoteAddr))

			strBody := ""
			strQueryParams := ""
			strQueryParams = c.QueryParams().Encode()
			body, errP := ioutil.ReadAll(c.Request().Body)
			if errP != nil {
				fmt.Printf("Error when read request body: %v", err)
				strBody = err.Error()
			} else {
				strBody = string(body)
			}

			span.SetAttributes(attribute.Key("Request Query Params").String(strQueryParams))
			span.SetAttributes(attribute.Key("Request Body").String(strBody))

			// Add the span to the request context so it can be used by other handlers
			c.SetRequest(c.Request().WithContext(trace.ContextWithSpan(c.Request().Context(), span)))

			// Call the next middleware/handler in the chain
			err := next(c)

			// Finish the span when the request is complete
			span.End()

			return err
		}
	})

	// Define some example routes
	ech.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello, world!"})
	})

	ech.POST("/foo", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "You posted to /foo"})
	})

	// Wrap the Echo router with otelhttp.NewHandler to instrument it with OpenTelemetry
	h := otelhttp.NewHandler(ech, "example/echo")

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))

}
