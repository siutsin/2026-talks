# Darwinian Concurrency Supervisor

A high-performance Wikipedia crawler that utilizes **Evolutionary Algorithms (GA)** and **PID Control** to autonomously optimize its own crawling strategy in real-time.

## Overview

The Darwinian Concurrency Supervisor is not just a simple worker pool; it's an adaptive system designed to find the optimal balance between throughput and reliability. It crawls Wikipedia pages, extracts links, and feeds them back into a queue, all while a "Darwinian Search" layer continuously experiments with different execution parameters (DNA) to maximize data ingestion speed while minimizing error rates.

## Key Features

- **Evolutionary Strategy (Genetic Algorithm):** Periodically mutates execution parameters like `MaxWorkers`, `BatchSize`, and `MaxRetries`. It selects the "fittest" strategy based on a scoring function that balances MB/s and success rate.
- **PID-Controlled Backpressure:** Uses a Proportional-Integral-Derivative (PID) controller reflex to immediately throttle worker activity if error rates (e.g., HTTP 429s or timeouts) spike, providing faster protection than the GA layer.
- **Dynamic Worker Pool:** Real-time adjustment of goroutine counts without restarting the engine.
- **Real-time Telemetry UI:** A graphical dashboard built with [Fyne](https://fyne.io/) that visualizes the current strategy "DNA", queue depth, ingestion rates, and execution reliability.
- **Wikipedia Link Extraction:** Specifically tuned to crawl Wikipedia's internal link structure.

## Architecture

- **Engine:** The core orchestrator managing workers, the queue, and the evolutionary loop.
- **Strategy:** Defines the "DNA" of the execution (Workers, Batching, Retries, Backoff, PID Gain).
- **Telemetry:** Atomic counters tracking successes, failures, bytes read, and latency.
- **Crawler:** HTTP fetcher with regex-based link extraction.
- **UI:** A Fyne-based observer that polls the engine's state for live visualization.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.25.5 or later recommended)
- System dependencies for Fyne (see [Fyne Setup](https://developer.fyne.io/started/))

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/caseylmanus/workerpool.git
   cd workerpool
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running the Application

To start the crawler and the observer UI:

```bash
go run .
```

The UI will appear, showing the initial strategy being seeded. Over time, you will see the "Supervisor Status" switch to "DARWINIAN SEARCH" as it begins its evolutionary optimization.

## Testing

Run the test suite to verify the engine and its components:

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details (or add one!).
