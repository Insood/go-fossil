#version 330

in vec2 fragTexCoord;
in vec4 fragLightClipPosition;
in vec3 fragWorldNormal;

uniform vec4 colDiffuse;
uniform sampler2D texture0;
uniform sampler2D texture2;
uniform vec3 lightDirection;
uniform float shadowBias;
uniform float shadowSlopeBias;
uniform float shadowDarkness;

out vec4 finalColor;

float sampleShadow(vec4 lightClipPosition, vec3 worldNormal) {
    vec3 projected = lightClipPosition.xyz / lightClipPosition.w;
    projected = projected * 0.5 + 0.5;

    bool outsideShadowMap =
        projected.x < 0.0 || projected.x > 1.0 ||
        projected.y < 0.0 || projected.y > 1.0 ||
        projected.z < 0.0 || projected.z > 1.0;

    if (outsideShadowMap) {
        return 0.0;
    }

    float closestDepth = texture(texture2, projected.xy).r;
    float lightAlignment = max(dot(normalize(worldNormal), normalize(-lightDirection)), 0.0);
    float bias = shadowBias + shadowSlopeBias * (1.0 - lightAlignment);

    return projected.z - bias > closestDepth ? shadowDarkness : 0.0;
}

void main() {
    vec4 albedo = texture(texture0, fragTexCoord);
    if (albedo.a <= 0.05) {
        discard;
    }

    vec3 baseColor = colDiffuse.rgb * albedo.rgb;
    float shadow = sampleShadow(fragLightClipPosition, fragWorldNormal);

    finalColor = vec4(baseColor * (1.0 - shadow), colDiffuse.a * albedo.a);
}
