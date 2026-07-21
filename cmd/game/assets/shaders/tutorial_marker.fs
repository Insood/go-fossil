#version 330

in vec3 fragWorldNormal;
in vec4 fragColor;

uniform vec4 colDiffuse;

out vec4 finalColor;

void main() {
    vec3 normal = normalize(fragWorldNormal);
    vec3 lightDirection = normalize(vec3(-0.35, 0.85, -0.40));

    float facing = max(dot(normal, lightDirection), 0.0);
    float facetShade = 0.50 + facing * 0.50;
    float sideHighlight = pow(1.0 - abs(normal.y), 2.0) * 0.35;

    vec3 baseColor = colDiffuse.rgb * fragColor.rgb;
    vec3 shadedColor = baseColor * facetShade + vec3(1.0, 0.28, 0.20) * sideHighlight;

    finalColor = vec4(shadedColor, colDiffuse.a * fragColor.a);
}
