#version 330

in vec3 vertexPosition;
in vec3 vertexNormal;
in vec2 vertexTexCoord;
in vec4 vertexColor;

uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 lightViewProjection;
uniform sampler2D texture2;
uniform float shadowNormalBias;
uniform float terrainCutoutDivotDepth;
uniform float terrainCutoutOverlayAlpha;

out vec4 fragColor;
out vec2 fragTexCoord;
out vec4 fragLightClipPosition;
out vec3 fragWorldNormal;

void main() {
    vec4 burnOverlay = texture(texture2, vertexTexCoord);
    float cutoutMask = 1.0 - step(0.1, abs(burnOverlay.a - terrainCutoutOverlayAlpha));
    vec3 displacedPosition = vertexPosition - vec3(0.0, terrainCutoutDivotDepth * cutoutMask, 0.0);

    vec4 worldPosition = matModel * vec4(displacedPosition, 1.0);
    vec3 worldNormal = normalize(mat3(matModel) * vertexNormal);
    vec4 shadowSamplePosition = vec4(worldPosition.xyz + worldNormal * shadowNormalBias, 1.0);

    fragColor = vertexColor;
    fragTexCoord = vertexTexCoord;
    fragLightClipPosition = lightViewProjection * shadowSamplePosition;
    fragWorldNormal = worldNormal;
    gl_Position = mvp * vec4(displacedPosition, 1.0);
}
