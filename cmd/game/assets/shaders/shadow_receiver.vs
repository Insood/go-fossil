#version 330

in vec3 vertexPosition;
in vec3 vertexNormal;
in vec2 vertexTexCoord;
in vec4 vertexColor;

uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 lightViewProjection;

out vec4 fragColor;
out vec2 fragTexCoord;
out vec4 fragLightClipPosition;
out vec3 fragWorldNormal;

void main() {
    vec4 worldPosition = matModel * vec4(vertexPosition, 1.0);

    fragColor = vertexColor;
    fragTexCoord = vertexTexCoord;
    fragLightClipPosition = lightViewProjection * worldPosition;
    fragWorldNormal = normalize(mat3(matModel) * vertexNormal);
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}
