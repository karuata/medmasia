param(
    [int]$Port = 8020
)

$ErrorActionPreference = "Stop"

python -m http.server $Port --bind 127.0.0.1
