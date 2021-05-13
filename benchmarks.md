


./eott --profile=cpu.prof -r 1.4 songs/Something\ Wild/

Time: Apr 30, 2021 at 12:17pm (BST)
Duration: 1.64mins, Total samples = 10.48s (10.62%)

      flat  flat%   sum%        cum   cum%
     590ms  5.63%  5.63%      600ms  5.73%  math.Round (partial-inline)
     480ms  4.58% 10.21%     4310ms 41.13%  main.run.func5
     440ms  4.20% 14.41%      440ms  4.20%  runtime.futex
     390ms  3.72% 18.13%      390ms  3.72%  <unknown>
     350ms  3.34% 21.47%      390ms  3.72%  syscall.Syscall
     320ms  3.05% 24.52%     1780ms 16.98%  runtime.findrunnable
     290ms  2.77% 27.29%     1030ms  9.83%  fmt.(*pp).doPrintf
     280ms  2.67% 29.96%      280ms  2.67%  runtime.memclrNoHeapPointers
     280ms  2.67% 32.63%      280ms  2.67%  runtime.nextFreeFast (inline)
     250ms  2.39% 35.02%      250ms  2.39%  github.com/jfreymuth/vorbis.imdct
     250ms  2.39% 37.40%     1210ms 11.55%  runtime.mallocgc
     210ms  2.00% 39.41%      210ms  2.00%  runtime.memmove
     190ms  1.81% 41.22%      350ms  3.34%  runtime.scanobject
     160ms  1.53% 42.75%     1000ms  9.54%  github.com/jfreymuth/oggvorbis.(*Reader).Read
     140ms  1.34% 44.08%      190ms  1.81%  git.lost.host/meutraa/eott/internal/score.(*DefaultScorer).Distance
     140ms  1.34% 45.42%      450ms  4.29%  runtime.checkTimers
     140ms  1.34% 46.76%      180ms  1.72%  sync.(*Pool).Get
     130ms  1.24% 48.00%      130ms  1.24%  [libpthread-2.32.so]
     130ms  1.24% 49.24%      660ms  6.30%  fmt.(*pp).printArg
     130ms  1.24% 50.48%      320ms  3.05%  github.com/jfreymuth/vorbis.(*residue).Decode
     120ms  1.15% 51.62%      310ms  2.96%  github.com/hajimehoshi/oto/internal/mux.(*Mux).Read
     120ms  1.15% 52.77%      180ms  1.72%  github.com/jfreymuth/vorbis.huffmanCode.Lookup (inline)
     120ms  1.15% 53.91%      120ms  1.15%  runtime.(*randomEnum).next (inline)
     120ms  1.15% 55.06%      120ms  1.15%  runtime.cgocall
     120ms  1.15% 56.20%      120ms  1.15%  runtime.epollwait

Time: Apr 30, 2021 at 2:39pm (BST)
Duration: 1.62mins, Total samples = 9.43s ( 9.68%)

      flat  flat%   sum%        cum   cum%
     0.61s  6.47%  6.47%      0.61s  6.47%  runtime.futex
     0.61s  6.47% 12.94%      0.65s  6.89%  syscall.Syscall
     0.37s  3.92% 16.86%      0.37s  3.92%  git.lost.host/meutraa/eott/internal/score.(*DefaultScorer).Distance
     0.34s  3.61% 20.47%      3.31s 35.10%  main.run.func5
     0.29s  3.08% 23.54%      0.49s  5.20%  runtime.scanobject
     0.26s  2.76% 26.30%      0.26s  2.76%  <unknown>
     0.26s  2.76% 29.06%      0.26s  2.76%  runtime.memmove
     0.24s  2.55% 31.60%      0.98s 10.39%  runtime.mallocgc
     0.22s  2.33% 33.93%      0.22s  2.33%  github.com/jfreymuth/vorbis.imdct
     0.20s  2.12% 36.06%      0.20s  2.12%  [libpthread-2.32.so]
     0.20s  2.12% 38.18%      1.65s 17.50%  runtime.findrunnable
     0.20s  2.12% 40.30%      0.20s  2.12%  runtime.nextFreeFast (inline)
     0.18s  1.91% 42.21%      0.18s  1.91%  runtime.memclrNoHeapPointers
     0.15s  1.59% 43.80%      0.98s 10.39%  fmt.(*pp).doPrintf
     0.14s  1.48% 45.28%      0.73s  7.74%  github.com/jfreymuth/vorbis.(*Decoder).decodePacket
     0.14s  1.48% 46.77%      0.16s  1.70%  runtime.cgocall
     0.13s  1.38% 48.14%      0.26s  2.76%  fmt.(*fmt).fmtInteger
     0.13s  1.38% 49.52%      0.87s  9.23%  github.com/jfreymuth/oggvorbis.(*Reader).Read
     0.12s  1.27% 50.80%      0.12s  1.27%  runtime.(*randomEnum).next (inline)
     0.12s  1.27% 52.07%      0.12s  1.27%  runtime.pMask.read (inline)
     0.11s  1.17% 53.23%      0.15s  1.59%  runtime.lock2
     0.11s  1.17% 54.40%      0.14s  1.48%  runtime.nanotime (inline)
     0.10s  1.06% 55.46%      0.10s  1.06%  runtime.read
     0.10s  1.06% 56.52%      0.13s  1.38%  sync.(*Pool).Get
     0.09s  0.95% 57.48%      0.69s  7.32%  fmt.(*pp).printArg