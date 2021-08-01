% processQuat2
% this one!!
% normalizes image size to file the square
clear all


startInd = 1;

pn = 'C:\Users\john\Documents\Arduino\ardpifi\letters'
cd(pn)
cd('O')

dirs = dir;
nDirs = size(dirs,1);

% subplot: m by n images
sp_m = floor(sqrt(nDirs));
sp_n = floor(nDirs/sp_m);

nDirs2 = sp_m*sp_n;

figure
for d = 3:nDirs2
    
    dn=(dirs(d).name)
    
    dn = 'O_20-32-50'
    cd(dn)
    
    fn = 'quaternion_data.txt';
    X=importdata(fn,'%s');
    
%     dataL = length(X)*0.99
     dataL = length(X);
    
    quat = [];
    for i = startInd:dataL
        quat(i-startInd+1,:) = str2num(cell2mat(X(i)));
    end
    
    dataL = floor(dataL-startInd+1);
    
    

    
    %% step 1.
    % continuously as we received quat data from the imu:
    %     for each new quat, calc rotation = 3x3 matrix
    
    
    %     quat2rotm_ = zeros(3,3,dataL);
    %     quat2rotm_JU
    
    
    projected = zeros(dataL,3);
    
    for n=1:dataL
        
        s = quat(n,1);
        x = quat(n,2);
        y = quat(n,3);
        z = quat(n,4);
        
        %         R = [   1-2*(y.^2+z.^2)   2*(x.*y-s.*z)   2*(x.*z+s.*y)
        %             2*(x.*y+s.*z) 1-2*(x.^2+z.^2)   2*(y.*z-s.*x)
        %             2*(x.*z-s.*y)   2*(y.*z+s.*x) 1-2*(x.^2+y.^2)   ];
        %
        %         quat2rotm_(:,:,n) = R;
        
        % we only need the first colum
        
        
        projected(n,1) = 1-2*(y^2+z^2)  ;
        projected(n,2) = 2*(x*y+s*z)  ;
        projected(n,3) =  2*(x*z-s*y)  ;
        
        
        
    end
    
    
    
    % %     % 2. convert each 3x3 matrix into 3x1 projecton
    %     projected = [];
    %     for n=1:dataL
    %         projected(n,:) = quat2rotm_(:,:,n) * [1 0 0]';  % first column
    %     end
    
    %     stop
    
    %% step 3. when we have stopped recording data: average of projected
    
    centre = mean(projected, 1);
    
    
    
    %% step 4: norm-centre
    
    %     centre_direction = centre' ./ norm(centre);
    
    norm_centre = sqrt(centre(1)^2+centre(2)^2+centre(3)^2);
    
    centre_direction = zeros(3,1);
    centre_direction(1) = centre(1)/norm_centre;
    centre_direction(2) = centre(2)/norm_centre;
    centre_direction(3) = centre(3)/norm_centre;
    
    
    % step 5: y direction:
    %     y_direction = (eye(3) - centre_direction*centre_direction')* [0 1 0]';
    
    centre_direction_sq_centre_col(1) = centre_direction(1)*centre_direction(2);
    centre_direction_sq_centre_col(2) = centre_direction(2)*centre_direction(2);
    centre_direction_sq_centre_col(3) = centre_direction(2)*centre_direction(3);
    
    eye_minus_cdscc = zeros(3,1);
    eye_minus_cdscc(1) = -1.0*centre_direction_sq_centre_col(1);
    eye_minus_cdscc(2) = 1.0-centre_direction_sq_centre_col(2);
    eye_minus_cdscc(3) = -1.0*centre_direction_sq_centre_col(3);
    
    
    %     y_direction = y_direction ./ norm(y_direction);
    
    norm_y_direction = sqrt(eye_minus_cdscc(1)^2+eye_minus_cdscc(2)^2+eye_minus_cdscc(3)^2);
    
    y_direction = zeros(3,1);
    y_direction(1) = eye_minus_cdscc(1) ./ norm_y_direction;
    y_direction(2) = eye_minus_cdscc(2) ./ norm_y_direction; 
%     y_direction(2) = 1.0;    % tends to unity
    y_direction(3) = eye_minus_cdscc(3) ./ norm_y_direction;
    
    
    %%     step 6: x_direction via cross product
    %     x_direction_cp = cross(centre_direction, y_direction);
    x_direction = zeros(3,1);
    %     cx = aybz ? azby
    %     cy = azbx ? axbz
    %     cz = axby ? aybx
    x_direction(1) = centre_direction(2)*y_direction(3) - centre_direction(3)*y_direction(2);
    x_direction(2) = centre_direction(3)*y_direction(1) - centre_direction(1)*y_direction(3); % very close to zero
    x_direction(3) = centre_direction(1)*y_direction(2) - centre_direction(2)*y_direction(1);
    
    
    %% step 7: x and y corrodinates:
    
    %     projected2 = [];
    %     for n=1:dataL
    %         projected2(n,:) = [x_direction y_direction]' * projected(n, :)';
    %     end
    %     x = projected2(:, 1);
    %     y = projected2(:, 2);
    
    x = zeros(dataL,1);
    y = zeros(dataL,1);
    
    pCP = 1.09;
     pCN = 1/pCP;
    for n=1:dataL
        
        x(n) = x_direction(1)* projected(n, 1) + x_direction(2)* projected(n, 2)  + x_direction(3)* projected(n, 3) ;
        y(n) = y_direction(1)* projected(n, 1) + y_direction(2)* projected(n, 2)  + y_direction(3)* projected(n, 3) ;
        
%          x(n+dataL) = x(n)*pCP;  
%          x(n+dataL+1) = x(n)*pCN;  
         
%          y(n+dataL) = y(n)*pCP;
%          y(n+2) = y(n)*pCN;
         
%          if n==1
%              y(n+2) = y(n);
%              x(n+2) = x(n);
%          else             
%              y(n-1) = y(n);
%              x(n-1) = x(n);
%          end
        
%         x(n) = x_direction(1)* projected(n, 1)  + x_direction(3)* projected(n, 3) ; % x_direction(2) is so small we can ignore it
%          y(n) = y_direction(1)* projected(n, 1) +  projected(n, 2)  + y_direction(3)* projected(n, 3) ; % y_direction(2) -> 1.0

    end
    
    dataL = length(x);
    
    
    % scale
    
    if(abs(min(x))>max(x))
        maxX = min(x);
    else
        maxX = max(x);
    end
        
    if(abs(min(y))>max(y))
        maxY = min(y);
    else
        maxY = max(y);
    end
    
    if abs(maxY)>abs(maxX)
        maxdim = maxY;
    else
        maxdim = maxX;
    end
     
    scaler = 1/abs(maxdim);
    
    x = x.*scaler;
    y = y.*scaler;
    
    scaler
    
    
    
    %%
    % convert vector into bitmap
        
    m_x = 28;   % pixels in square
    m_y = m_x;
    m = zeros(m_x,m_x);
    
    x_int = round(x*m_x/2);
    y_int = round(y*m_y/2);
    
 
    for n=1:dataL
        x_int_d = x_int(n)+m_x/2 + 1;
        y_int_d = y_int(n)+m_y/2 + 1;
        
        m(x_int_d,y_int_d) = 1;
    end
    
        
%     subplot(sp_m,sp_n,d-2)
        figure
    I = mat2gray(m);
    imshow(I)    
    
    stop
    
        
    cd('C:\Users\john\Documents\Arduino\ardpifi\letters')
    fn = [int2str(d) '_O.bmp']
    imwrite(I,fn)
    
    
    cd(pn)
cd('O')
    %     hold on;
    
%     cd ..
    
    
end