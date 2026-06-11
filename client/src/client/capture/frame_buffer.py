import asyncio

class FBuf:
    def __init__(self, val_type: type, maxsize: int = 3):
        self.val_type: type = val_type
        self._maxsize: int = maxsize
        self.q = asyncio.Queue(maxsize=self._maxsize)

    async def put(self, value):
        if not isinstance(value, self.val_type):
            raise TypeError(f"В очередь типа {self.val_type.__name__} нельзя положить элемент с типом {type(value).__name__}")
        
        if self.q.full():
            try:
                self.q.get_nowait()
            except asyncio.QueueEmpty:
                pass
        
        await self.q.put(value)

    async def get(self):
        return await self.q.get()
