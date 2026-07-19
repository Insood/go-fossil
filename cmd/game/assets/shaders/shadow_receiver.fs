#version 330

in vec4 fragColor;
in vec2 fragTexCoord;
in vec4 fragLightClipPosition;
in vec3 fragWorldNormal;

uniform vec4 colDiffuse;
uniform sampler2D texture0;
uniform sampler2D texture1;
uniform sampler2D texture2;
uniform sampler2D shadowMap;
uniform vec3 lightDirection;
uniform float shadowBias;
uniform float shadowSlopeBias;
uniform float shadowDarkness;
uniform float slopeShadeStrength;

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

    float closestDepth = texture(shadowMap, projected.xy).r;
    float lightAlignment = max(dot(normalize(worldNormal), normalize(-lightDirection)), 0.0);
    float bias = shadowBias + shadowSlopeBias * (1.0 - lightAlignment);

    return projected.z - bias > closestDepth ? shadowDarkness : 0.0;
}

void main() {
    vec4 groundLayer = texture(texture0, fragTexCoord);
    vec4 artifactLayer = texture(texture1, fragTexCoord);
    vec4 burnOverlay = texture(texture2, fragTexCoord);

    vec3 baseColor = colDiffuse.rgb * groundLayer.rgb;
    baseColor = mix(baseColor, artifactLayer.rgb, artifactLayer.a);
    baseColor = mix(baseColor, burnOverlay.rgb, burnOverlay.a);

    float shadow = sampleShadow(fragLightClipPosition, fragWorldNormal);
    vec3 worldNormal = normalize(fragWorldNormal);
    float slopeShade = (1.0 - clamp(worldNormal.y, 0.0, 1.0)) * slopeShadeStrength;
    vec3 shadedColor = baseColor * (1.0 - shadow) * (1.0 - slopeShade);

    float alpha = colDiffuse.a * groundLayer.a;
    finalColor = vec4(shadedColor, alpha);
}
