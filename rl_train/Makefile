freeze-env:
	@echo "Freezing environment..."
	@venv/bin/pip freeze > requirements.txt

install:
	@echo "Installing dependencies..."
	@venv/bin/pip install -r requirements.txt

activate-env:
	@echo "Activating environment..."
	venv/bin/activate

venv:
	@echo "Creating virtual environment..."
	@python3 -m venv venv

train:
	@echo "Training model..."
	@venv/bin/python3 train.py
