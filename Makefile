build-game:
	@echo "Building game..."
	@mkdir -p build
	@cd build && go build -o game ../server/run_game_server.go
	@cp -r assets build

train:
	@echo "Training..."
	@cd rl_train && make activate-env && make train

run-game:
	@echo "Running game..."
	@cd build && ./game
