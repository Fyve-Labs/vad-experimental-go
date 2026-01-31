# VAD Experimental

Usage:

```bash
bash run.sh
```

Output:

```text
>Started. Please speak. Press ctrl + C  to exit
2026/01/30 20:30:29.868798 Detected speech
2026/01/30 20:30:30.661034 Saved to seg-0-0.86-seconds.wav
2026/01/30 20:30:30.661082 Duration: 0.86 seconds
2026/01/30 20:30:30.661094 ----------
2026/01/30 20:30:37.468778 Detected speech
2026/01/30 20:30:39.481753 Saved to seg-1-2.08-seconds.wav
2026/01/30 20:30:39.481804 Duration: 2.08 seconds
2026/01/30 20:30:39.481815 ----------
2026/01/30 20:30:40.828849 Detected speech
2026/01/30 20:30:43.530990 Saved to seg-2-2.75-seconds.wav
2026/01/30 20:30:43.531033 Duration: 2.75 seconds
2026/01/30 20:30:43.531044 ----------
```

Raspberry Pi 2 W resource consumption (fresh OS installation)

```txt
    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND                                                                                                      
   1125 viet      20   0 1781092  36632  20108 S  14.2   8.6   3:02.66 vad                                                                                                          
```