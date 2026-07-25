from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    BatchSpanProcessor
)
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import (
    OTLPSpanExporter
)
from opentelemetry.instrumentation.fastapi import (
    FastAPIInstrumentor
)


def init_observability(app):
    provider = TracerProvider()
    exporter = OTLPSpanExporter(endpoint="otel-collector:4317", insecure=True)
    provider.add_span_processor(
        BatchSpanProcessor(exporter)
    )
    trace.set_tracer_provider(provider)
    FastAPIInstrumentor.instrument_app(app)
