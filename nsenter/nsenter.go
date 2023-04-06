package nsenter

/*
#define _GNU_SOURCE
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sched.h>
#include <errno.h>

__attribute__((constructor)) void enter_namespace(void)
{
    char *docker_pid = getenv("minidocker_pid");
    if (docker_pid)
    {
        printf("get docker pid : %s\n", docker_pid);
    }
    else
    {
        printf("missing docker pid");
        return;
    }

    char *docker_cmd = getenv("minidocker_cmd");
    if (docker_cmd)
    {
        printf("get docker cmd : %s\n", docker_cmd);
    }
    else
    {
        printf("missing docker cmd");
        return;
    }

    char nspath[1024];
    char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};
    for (size_t i = 0; i < 5; i++)
    {
        sprintf(nspath, "/proc/%s/ns/%s", docker_pid, namespaces[i]);
        int fd = open(nspath, O_RDONLY);
        if (setns(fd, 0) == -1)
        {
            printf("setns on %s failed : %s\n", namespaces[i], strerror(errno));
        }
        else
        {
            printf("setns on %s succeeded\n", namespaces[i]);
        }
        close(fd);
    }

    int res = system(docker_cmd);
	exit(0);
    return;
}
*/
import "C"
