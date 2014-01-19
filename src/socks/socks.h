#include <stdint.h>
#include <sys/types.h>
#include <stddef.h>

int createDeviceBoundUDPSocket(uint32_t, uint16_t, const char*);

ssize_t writeToUDP(int, void*, size_t, uint32_t, uint16_t);

ssize_t recvFromUDP(int, void *, size_t, uint32_t *, uint16_t *);
