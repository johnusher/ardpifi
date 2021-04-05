
clear all
ca

cd('C:\Users\john\Dropbox\skitz\letters\M\')

dirs = dir;
nDirs = size(dirs,1);

for d = 3:nDirs
    
    dn=(dirs(d).name)
    cd(dn)
    
    fn = 'euler_data.txt'
    X=importdata(fn,'%s');
    
    % X = cellfun(X,3);
    
    bsf = 0;    % bearingScaleFlag
    
    for i = 26:length(X)*0.9
        b(i,:) = str2num(cell2mat(X(i)));
    end
    x = b(:,1);
    y = b(:,2);
    z = b(:,3);
    
    x=unwrap(x,80);
    
    
%     figure;plot(b)
%     legend('x','y','z')
    
    
    %  dataout =  iomega(datain, dt, datain_type, dataout_type)
    
    % dataoutX =  iomega(x, 0.1, 3, 1);
    % dataoutY =  iomega(y, 0.1, 3, 1);
    % dataoutZ =  iomega(z, 0.1, 3, 1);
    
    % figure;plot(x);hold on;plot(dataoutX,'x'); title('bearin')
    % figure;plot(y);hold on;plot(dataoutY,'x'); title('roll')
    % figure;plot(z);hold on;plot(dataoutZ,'x'); title('tilt')
    
    
%     figure;plot(x); title('bearin')
%     figure;plot(y); title('roll')
%     figure;plot(z); title('tilt')

% y(150:155)
    
    hold on;plot(y)
    
    cd ..
    
    
end