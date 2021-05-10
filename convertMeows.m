% convertMeows.m

ca
inPN = '\meows_In';
outPN = '\meows_Out\';
cd(inPN)

outFS = 48e3;

filesWavs = dir('*.wav');
filesMp3s = dir('*.mp3');

nWavFiles = size(filesWavs,1)-0;
nMP3Files = size(filesMp3s,1)-0;
totFiles = nWavFiles+nMP3Files;

smT = 0.1*outFS;
inWin = hanning(smT);
inWin = inWin(1:smT/2);

dcblker = dsp.DCBlocker;

for n=1:totFiles
    fn = filesWavs(n).name;
    [in,fsI] = audioread(fn);
    in = in(:,1);
    in = dcblker(in);
    in = in-mean(in);
    in = in./(1.2*max(abs(in)));
    in(1:smT/2) = in(1:smT/2).* inWin;
    out = resample(in,outFS,fsI);
    
    sn = [outPN 'meow_' int2str(n) '.wav'];
    audiowrite(sn,out,outFS);    
    
    
end


