#include <netinet/in.h>
#include <errno.h>
#include <stdlib.h>
#include <unistd.h>
#include <netdb.h>
#include <string.h>
#include <stdio.h>
#include "socks.h"

int createDeviceBoundUDPSocket(uint32_t sip, uint16_t sport, const char* bind_dev) {
    printf("bind_dev = %s", bind_dev);
   int s = socket(AF_INET, SOCK_DGRAM, 0);   
   int result;
   struct sockaddr_in my_ip_addr;

   if (s < 0) {
      perror("socket");
      return s;
   }

   memset(&my_ip_addr, 0, sizeof(my_ip_addr));
   
   my_ip_addr.sin_family = AF_INET;
   my_ip_addr.sin_addr.s_addr = htonl(sip);
   my_ip_addr.sin_port = htons(sport);
   
   /*
   result = bind(s, (struct sockaddr*)(&my_ip_addr), sizeof(my_ip_addr));
   if (result < 0) {
      perror("Error in bind");
      return result;
   }
   */
   
   if (bind_dev) {
      // Bind to specific device.
      if (setsockopt(s, SOL_SOCKET, SO_BINDTODEVICE,
                     bind_dev, strlen(bind_dev) + 1)) {
         perror("Error binding to device");
         return -1;
      }
   }

   return s;
}

ssize_t writeToUDP(int fd, void *buf, size_t buflen, uint32_t dip, uint16_t dport) {
    struct sockaddr_in sa;
    sa.sin_addr.s_addr = htonl(dip);
    sa.sin_family = AF_INET;
    sa.sin_port = htons(dport);

    return sendto(fd, buf, buflen, 0, &sa, sizeof(sa));
}

ssize_t recvFromUDP(int fd, void *buf, size_t buflen, uint32_t *from_ip_buf, uint16_t *from_port_buf) {
    // IPv6 support needs sockaddr_storage to then cast to in or in6
    struct sockaddr_in sa;
    // why is it a pointer to a socklen_t?
    socklen_t sa_len = sizeof sa;

    ssize_t res = recvfrom(fd, buf, buflen, 0, &sa, &sa_len);
    //ssize_t res = recvfrom(fd, buf, buflen, 0, &sa, sizeof(sa));
    //int * crash = NULL; if (res < 0) { *crash = errno; } // force crash

    // pass in src ip and port out of band
    *from_ip_buf = ntohl(sa.sin_addr.s_addr);
    *from_port_buf = ntohs(sa.sin_port);

    return res;
}
