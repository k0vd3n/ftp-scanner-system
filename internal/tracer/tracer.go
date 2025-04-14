package tracer

// import (
// 	"fmt"
// 	"ftp-scanner_try2/config"
// 	"io"
// 	"log"

// 	"github.com/uber/jaeger-client-go"
// 	jaegercfg "github.com/uber/jaeger-client-go/config"

// 	"github.com/opentracing/opentracing-go"
// )

// func InitJaeger(cfg config.JaegerConfig) (opentracing.Tracer, io.Closer) {
// 	agentEndpoint := fmt.Sprintf("%s:%s", cfg.AgentHost, cfg.AgentPort)
// 	jCfg := jaegercfg.Configuration{
// 		ServiceName: cfg.ServiceName,
// 		Sampler: &jaegercfg.SamplerConfig{
// 			Type:  cfg.Sampler.Type,
// 			Param: cfg.Sampler.Param,
// 		},
// 		Reporter: &jaegercfg.ReporterConfig{
// 			LogSpans:           cfg.Reporter.LogSpans,
// 			LocalAgentHostPort: agentEndpoint,
// 		},
// 	}
// 	tracer, closer, err := jCfg.NewTracer(
// 		jaegercfg.Logger(jaeger.StdLogger),
// 	)
// 	if err != nil {
// 		log.Fatalf("Не удалось инициализировать Jaeger трейсер: %v", err)
// 	}
// 	opentracing.SetGlobalTracer(tracer)
// 	return tracer, closer
// }
