import time
from datetime import datetime, timezone

if __name__ == "__main__":
    # Inject credentails to application
    print("Running entrypoint of orderbook")

    while True:
        print(f"Running... {datetime.now(timezone.utc)}")
        time.sleep(5)
