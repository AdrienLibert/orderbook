from drgn.env import config_from_env

if __name__ == "__main__":
    print("Running entrypoint of orderbook")
    config_from_env()

    while True:
        print(f"Running... {datetime.now(timezone.utc)}")
        time.sleep(5)
