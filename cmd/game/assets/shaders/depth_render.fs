#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform sampler2D texture0;
uniform float nearPlane;
uniform float farPlane;
uniform float isOrthographic;

float linearizePerspectiveDepth(float depth) {
    float z = depth*2.0 - 1.0;
    return (2.0*nearPlane*farPlane)/(farPlane + nearPlane - z*(farPlane - nearPlane));
}

void main() {
    float depth = texture(texture0, fragTexCoord).r;
    float linearDepth = mix(
        linearizePerspectiveDepth(depth),
        nearPlane + depth*(farPlane - nearPlane),
        isOrthographic
    );
    float normalizedDepth = clamp((linearDepth - nearPlane)/(farPlane - nearPlane), 0.0, 1.0);
    finalColor = vec4(vec3(normalizedDepth), 1.0);
}
