import ctypes

lib = ctypes.CDLL("../cmd/server/libsonar.so")

lib.StartServerFromEnv.restype = ctypes.c_int
lib.StartServer.argtypes = (ctypes.c_int, ctypes.c_int)
lib.StartServer.restype = ctypes.c_int

# start server using environment variables
result = lib.StartServerFromEnv()
print(f"StartServerFromEnv returned: {result}")

# start server manually
result = lib.StartServer(8080, 1)
print(f"StartServer returned: {result}")
