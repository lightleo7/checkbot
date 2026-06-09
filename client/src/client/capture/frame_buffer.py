from queue import Queue, Full, Empty

class FBuf:

    def __init__(self, val_type: type, maxsize: int = 3):
        self.val_type: type = val_type
        self._maxsize: int = maxsize
        self.q = Queue(maxsize=self._maxsize)

    def put(self, value):
        if not isinstance(value, self.val_type):
            raise TypeError(f"В очередь типа {self.val_type.__name__} нельзя положить элемент с типом  {type(value).__name__}")
        try:
            self.q.put_nowait(value)
        except Full:
            self.q.get_nowait()
            self.q.put_nowait(value)

    def get(self):
        try:
            return self.q.get()
        except Empty:
            return None
