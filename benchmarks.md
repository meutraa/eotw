


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

Showing nodes accounting for 4520ms, 59.01% of 7660ms total
Dropped 150 nodes (cum <= 38.30ms)
Showing top 25 nodes out of 174
      flat  flat%   sum%        cum   cum%
     490ms  6.40%  6.40%      490ms  6.40%  runtime.futex
     490ms  6.40% 12.79%      540ms  7.05%  syscall.Syscall
     360ms  4.70% 17.49%     1780ms 23.24%  runtime.findrunnable
     240ms  3.13% 20.63%      260ms  3.39%  github.com/jfreymuth/vorbis.imdct
     230ms  3.00% 23.63%      540ms  7.05%  github.com/jfreymuth/vorbis.(*residue).Decode
     210ms  2.74% 26.37%      310ms  4.05%  github.com/jfreymuth/vorbis.huffmanCode.Lookup (inline)
     190ms  2.48% 28.85%      570ms  7.44%  fmt.(*pp).printArg
     180ms  2.35% 31.20%      180ms  2.35%  <unknown>
     180ms  2.35% 33.55%      180ms  2.35%  runtime.memclrNoHeapPointers
     160ms  2.09% 35.64%      160ms  2.09%  runtime.madvise
     150ms  1.96% 37.60%      150ms  1.96%  runtime.cgocall
     140ms  1.83% 39.43%     1400ms 18.28%  github.com/jfreymuth/oggvorbis.(*Reader).Read
     140ms  1.83% 41.25%     1230ms 16.06%  github.com/jfreymuth/vorbis.(*Decoder).decodePacket
     140ms  1.83% 43.08%      140ms  1.83%  runtime.epollwait
     140ms  1.83% 44.91%      140ms  1.83%  runtime.nextFreeFast (inline)
     130ms  1.70% 46.61%      130ms  1.70%  runtime.(*randomEnum).next (inline)
     130ms  1.70% 48.30%      130ms  1.70%  runtime.memmove
     120ms  1.57% 49.87%      280ms  3.66%  runtime.checkTimers
     120ms  1.57% 51.44%      190ms  2.48%  runtime.scanobject
     100ms  1.31% 52.74%      690ms  9.01%  fmt.(*pp).doPrintf
     100ms  1.31% 54.05%      290ms  3.79%  github.com/hajimehoshi/oto/internal/mux.(*Mux).Read
     100ms  1.31% 55.35%      100ms  1.31%  github.com/jfreymuth/vorbis.(*bitReader).Read1 (inline)
     100ms  1.31% 56.66%      650ms  8.49%  runtime.mallocgc
     100ms  1.31% 57.96%      100ms  1.31%  runtime.read
      80ms  1.04% 59.01%      130ms  1.70%  fmt.(*fmt).fmtInteger

Showing nodes accounting for 2680ms, 68.37% of 3920ms total
Dropped 70 nodes (cum <= 19.60ms)
Showing top 25 nodes out of 172
      flat  flat%   sum%        cum   cum%
     400ms 10.20% 10.20%      400ms 10.20%  runtime.futex
     260ms  6.63% 16.84%      430ms 10.97%  github.com/jfreymuth/vorbis.(*residue).Decode
     250ms  6.38% 23.21%      970ms 24.74%  runtime.findrunnable
     140ms  3.57% 26.79%      140ms  3.57%  <unknown>
     130ms  3.32% 30.10%      140ms  3.57%  github.com/jfreymuth/vorbis.imdct
     130ms  3.32% 33.42%      140ms  3.57%  syscall.Syscall
     120ms  3.06% 36.48%      120ms  3.06%  runtime.nanotime (inline)
     120ms  3.06% 39.54%      120ms  3.06%  runtime.unlock2
     100ms  2.55% 42.09%      300ms  7.65%  github.com/hajimehoshi/oto/internal/mux.(*Mux).Read
     100ms  2.55% 44.64%     1010ms 25.77%  github.com/jfreymuth/oggvorbis.(*Reader).Read
      90ms  2.30% 46.94%       90ms  2.30%  [libpthread-2.32.so]
      90ms  2.30% 49.23%      170ms  4.34%  github.com/jfreymuth/vorbis.huffmanCode.Lookup (inline)
      90ms  2.30% 51.53%      120ms  3.06%  runtime.cgocall
      80ms  2.04% 53.57%       80ms  2.04%  github.com/jfreymuth/vorbis.(*bitReader).Read1 (inline)
      70ms  1.79% 55.36%      880ms 22.45%  github.com/jfreymuth/vorbis.(*Decoder).decodePacket
      70ms  1.79% 57.14%      260ms  6.63%  runtime.checkTimers
      60ms  1.53% 58.67%       60ms  1.53%  github.com/jfreymuth/vorbis.(*Decoder).inverseCoupling
      50ms  1.28% 59.95%     1090ms 27.81%  github.com/faiface/beep.(*Mixer).Stream
      50ms  1.28% 61.22%       50ms  1.28%  runtime.(*randomEnum).next (inline)
      50ms  1.28% 62.50%       50ms  1.28%  runtime.lock2
      50ms  1.28% 63.78%       50ms  1.28%  runtime.madvise
      50ms  1.28% 65.05%       50ms  1.28%  runtime.memclrNoHeapPointers
      50ms  1.28% 66.33%       50ms  1.28%  runtime.read
      40ms  1.02% 67.35%     1250ms 31.89%  github.com/faiface/beep/speaker.update
      40ms  1.02% 68.37%       40ms  1.02%  runtime.epollwait

      Showing nodes accounting for 2660ms, 67.34% of 3950ms total
Dropped 76 nodes (cum <= 19.75ms)
Showing top 25 nodes out of 172
      flat  flat%   sum%        cum   cum%
     470ms 11.90% 11.90%      470ms 11.90%  runtime.futex
     260ms  6.58% 18.48%      260ms  6.58%  [libspa-audioconvert.so]
     210ms  5.32% 23.80%      210ms  5.32%  [libpthread-2.32.so]
     180ms  4.56% 28.35%      180ms  4.56%  github.com/jfreymuth/vorbis.imdct
     150ms  3.80% 32.15%      920ms 23.29%  runtime.findrunnable
     120ms  3.04% 35.19%      870ms 22.03%  github.com/jfreymuth/oggvorbis.(*Reader).Read
     110ms  2.78% 37.97%      420ms 10.63%  github.com/hajimehoshi/oto/internal/mux.(*Mux).Read
     100ms  2.53% 40.51%      670ms 16.96%  github.com/jfreymuth/vorbis.(*Decoder).decodePacket
      90ms  2.28% 42.78%      110ms  2.78%  syscall.Syscall
      80ms  2.03% 44.81%      140ms  3.54%  github.com/jfreymuth/vorbis.huffmanCode.Lookup (inline)
      80ms  2.03% 46.84%       80ms  2.03%  runtime.(*randomEnum).next (inline)
      80ms  2.03% 48.86%      100ms  2.53%  runtime.lock2
      80ms  2.03% 50.89%       80ms  2.03%  runtime.nanotime (inline)
      70ms  1.77% 52.66%      210ms  5.32%  github.com/jfreymuth/vorbis.(*residue).Decode
      70ms  1.77% 54.43%      120ms  3.04%  runtime.cgocall
      70ms  1.77% 56.20%      100ms  2.53%  runtime.scanobject
      60ms  1.52% 57.72%       60ms  1.52%  github.com/jfreymuth/vorbis.(*Decoder).inverseCoupling
      60ms  1.52% 59.24%       60ms  1.52%  github.com/jfreymuth/vorbis.(*bitReader).Read1 (inline)
      50ms  1.27% 60.51%       50ms  1.27%  [libc-2.32.so]
      50ms  1.27% 61.77%       50ms  1.27%  runtime.epollwait
      50ms  1.27% 63.04%       50ms  1.27%  runtime.madvise
      50ms  1.27% 64.30%       50ms  1.27%  runtime.runqgrab
      40ms  1.01% 65.32%       40ms  1.01%  github.com/jfreymuth/oggvorbis.crcUpdate
      40ms  1.01% 66.33%       40ms  1.01%  github.com/jfreymuth/vorbis.renderLine
      40ms  1.01% 67.34%       40ms  1.01%  runtime.memclrNoHeapPointers