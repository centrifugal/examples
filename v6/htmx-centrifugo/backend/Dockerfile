FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY main.py .
COPY database.py .
COPY templates/ ./templates/

# Expose port
EXPOSE 4000

# Start server
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "4000"]
