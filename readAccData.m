ca

// cd('C:\Users\john\Dropbox\skitz')
fn = 'acc_data.txt'
X=importdata(fn,'%s');

% X = cellfun(X,3);

for i = 1:length(X)
   b(i,:) = str2num(cell2mat(X(i))); 
end

figure;plot(b)