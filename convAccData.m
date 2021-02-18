% process accelerometer data.
% iomega from https://de.mathworks.com/matlabcentral/answers/21700-finding-the-velocity-from-displacement

ca


fn = 'acc_data.txt'
X=importdata(fn,'%s');

% X = cellfun(X,3);

for i = 1:length(X)*0.9
    b(i,:) = str2num(cell2mat(X(i)));
    if i>1
        
        for n=1:3
            if(abs(b(i,n)))>1.5*abs(b(i-1,n))
                'bad data'
                b(i,n) = b(i-1,n);
                
            end
        end
    end
end

figure;plot(b)

% b = b-9.39;

x = b(:,1);
y = b(:,2);
z = b(:,3);


%  dataout =  iomega(datain, dt, datain_type, dataout_type)

dataoutX =  iomega(x, 0.1, 3, 1);
dataoutY =  iomega(z, 0.1, 3, 1);

figure;plot(dataoutX)
figure;plot(dataoutY)
figure;plot(dataoutX,dataoutY)
