package main

func main() {
	fetcher := &HTTPFetcher{}
	engine := NewEngine(fetcher)
	engine.Start()

	ui := NewUI(engine, nil)
	ui.Run()
}
