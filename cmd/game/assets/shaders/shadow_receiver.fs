#version 330

in vec4 fragColor;
in vec2 fragTexCoord;
in vec4 fragLightClipPosition;

uniform vec4 colDiffuse;
uniform sampler2D texture0;
uniform sampler2D shadowMap;
uniform float shadowBias;
uniform float shadowDarkness;

out vec4 finalColor;

float sampleShadow(vec4 lightClipPosition) {
    vec3 projected = lightClipPosition.xyz / lightClipPosition.w;
    projected = projected * 0.5 + 0.5;

    bool outsideShadowMap =
        projected.x < 0.0 || projected.x > 1.0 ||
        projected.y < 0.0 || projected.y > 1.0 ||
        projected.z < 0.0 || projected.z > 1.0;

    if (outsideShadowMap) {
        return 0.0;
    }

    float closestDepth = texture(shadowMap, projected.xy).r;
    return projected.z - shadowBias > closestDepth ? shadowDarkness : 0.0;
}

void main() {
    vec4 albedo = texture(texture0, fragTexCoord);
    vec3 baseColor = colDiffuse.rgb * fragColor.rgb * albedo.rgb;
    float shadow = sampleShadow(fragLightClipPosition);
    vec3 shadedColor = baseColor * (1.0 - shadow);

    finalColor = vec4(shadedColor, colDiffuse.a * fragColor.a * albedo.a);
}
