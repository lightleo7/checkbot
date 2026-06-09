import constants

def get_start_time() -> int:
    with open(constants.START_TIME_FILE, "r") as f:
        return int(f.readline())
