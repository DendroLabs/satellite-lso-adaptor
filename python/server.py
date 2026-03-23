"""
Satellite orbital context HTTP service.

Lightweight HTTP server that exposes the orbital calculation engine
to the Go adaptor. Runs as a sidecar service alongside the main Go process.
"""

import json
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs
import sys
import os

sys.path.insert(0, os.path.dirname(__file__))
from orbital.coverage import check_coverage
from orbital.latency import estimate_one_way_latency
from orbital.constellation import LatLon


class OrbitalHandler(BaseHTTPRequestHandler):

    def do_GET(self):
        parsed = urlparse(self.path)
        params = parse_qs(parsed.query)

        routes = {
            "/coverage": self.handle_coverage,
            "/latency": self.handle_latency,
            "/health": self.handle_health,
        }

        handler = routes.get(parsed.path)
        if handler is None:
            self.send_json(404, {"error": "not found"})
            return

        try:
            handler(params)
        except (KeyError, ValueError) as e:
            self.send_json(400, {"error": str(e)})
        except Exception as e:
            self.send_json(500, {"error": str(e)})

    def handle_coverage(self, params):
        lat_a = float(params["latA"][0])
        lon_a = float(params["lonA"][0])
        lat_z = float(params["latZ"][0])
        lon_z = float(params["lonZ"][0])

        result = check_coverage(lat_a, lon_a, lat_z, lon_z)
        self.send_json(200, result)

    def handle_latency(self, params):
        lat_a = float(params["latA"][0])
        lon_a = float(params["lonA"][0])
        lat_z = float(params["latZ"][0])
        lon_z = float(params["lonZ"][0])

        a = LatLon(lat_a, lon_a)
        z = LatLon(lat_z, lon_z)
        result = estimate_one_way_latency(a, z)
        self.send_json(200, result)

    def handle_health(self, params):
        self.send_json(200, {"status": "ok", "service": "orbital-engine"})

    def send_json(self, status, data):
        body = json.dumps(data).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        # Structured logging to match Go service
        print(json.dumps({
            "service": "orbital-engine",
            "method": args[0] if args else "",
            "path": args[1] if len(args) > 1 else "",
            "status": args[2] if len(args) > 2 else "",
        }))


def main():
    port = int(os.environ.get("SATELLITE_SERVICE_PORT", "8090"))
    server = HTTPServer(("0.0.0.0", port), OrbitalHandler)
    print(json.dumps({"msg": "orbital engine starting", "port": port}))
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print(json.dumps({"msg": "orbital engine shutting down"}))
        server.shutdown()


if __name__ == "__main__":
    main()
