import time
import constants
def init():
    constants.ROOT_DATA_PATH.mkdir(exist_ok=True)
    with open(constants.START_TIME_FILE, "w") as f:
        f.write(str(int(time.time()*1000)))
