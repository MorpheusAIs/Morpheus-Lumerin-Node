dev:
	docker compose up ollama &
	local-run-proxy-router &
	cd ui-desktop && npm run dev

local-run-proxy-router:
	docker compose up --build proxy-router