#version 330

in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec4 vertexColor;

uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 lightViewProjection;

out vec4 fragColor;
out vec4 fragLightClipPosition;

void main() {
    vec4 worldPosition = matModel * vec4(vertexPosition, 1.0);

    fragColor = vertexColor;
    fragLightClipPosition = lightViewProjection * worldPosition;
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}
