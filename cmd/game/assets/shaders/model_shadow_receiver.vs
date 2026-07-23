#version 330

in vec3 vertexPosition;
in vec3 vertexNormal;
in vec2 vertexTexCoord;

uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 lightViewProjection;
uniform float shadowNormalBias;

out vec2 fragTexCoord;
out vec4 fragLightClipPosition;
out vec3 fragWorldNormal;

void main() {
    vec4 worldPosition = matModel * vec4(vertexPosition, 1.0);
    vec3 worldNormal = normalize(mat3(matModel) * vertexNormal);
    vec4 shadowSamplePosition = vec4(worldPosition.xyz + worldNormal * shadowNormalBias, 1.0);

    fragTexCoord = vertexTexCoord;
    fragLightClipPosition = lightViewProjection * shadowSamplePosition;
    fragWorldNormal = worldNormal;
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}
