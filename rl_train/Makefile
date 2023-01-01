freeze-env:
	@echo "Freezing environment..."
	@pip freeze > requirements.txt

install:
	@echo "Installing dependencies..."
	@pip install -r requirements.txt

activate-env:
	@echo "Activating environment..."
	venv/bin/activate

venv:
	@echo "Creating virtual environment..."
	@python3 -m venv venv

train:
	@echo "Training model..."
	@python3 train.py
