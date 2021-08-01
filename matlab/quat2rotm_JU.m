
% function R = quat2rotm( quat )

% quat2rotm_JU

%QUAT2ROTM Convert quaternion to rotation matrix
%   R = QUAT2ROTM(quat) converts a unit quaternion, quat, into an orthonormal
%   rotation matrix, R. The input, quat, is an N-by-4 matrix containing N quaternions.
%   Each quaternion represents a 3D rotation and is of the form quat = [w x y z],
%   with a scalar number as the first value. Each element of quat must be a real number.
%   The output, R, is an 3-by-3-by-N matrix containing N rotation matrices.
%
%   Example:
%      % Convert a quaternion to rotation matrix
%      quat = [0.7071 0.7071 0 0];
%      R = quat2rotm(quat)
%
%   See also rotm2quat



% robotics.internal.validation.validateNumericMatrix(quat, 'quat2rotm', 'quat', ...
%     'ncols', 4);

% % Normalize and transpose the quaternions
% quat = robotics.internal.normalizeRows(quat).';

% quat = quat';
% 
% % Reshape the quaternions in the depth dimension
% quatRS = reshape(quat,[4 1 size(quat,2)]);
% 
% s = quatRS(1,1,:);
% x = quatRS(2,1,:);
% y = quatRS(3,1,:);
% z = quatRS(4,1,:);
% 
% R = [   1-2*(y.^2+z.^2)   2*(x.*y-s.*z)   2*(x.*z+s.*y)
%     2*(x.*y+s.*z) 1-2*(x.^2+z.^2)   2*(y.*z-s.*x)
%     2*(x.*z-s.*y)   2*(y.*z+s.*x) 1-2*(x.^2+y.^2)   ];

for n=1:dataL
    
    s = quat(n,1);
    x = quat(n,2);
    y = quat(n,3);
    z = quat(n,4);
    
    R = [   1-2*(y.^2+z.^2)   2*(x.*y-s.*z)   2*(x.*z+s.*y)
    2*(x.*y+s.*z) 1-2*(x.^2+z.^2)   2*(y.*z-s.*x)
    2*(x.*z-s.*y)   2*(y.*z+s.*x) 1-2*(x.^2+y.^2)   ];

    quat2rotm_(:,:,n) = R;
    
end

