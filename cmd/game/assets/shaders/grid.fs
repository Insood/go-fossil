#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform vec4 colDiffuse;
uniform float gridCells;
uniform float lineWidth;

out vec4 finalColor;

void main() {
    vec2 scaled = fragTexCoord * gridCells;
    vec2 grid = min(fract(scaled), 1.0 - fract(scaled));
    float dist = min(grid.x, grid.y);

    float lineMask = 1.0 - smoothstep(0.0, lineWidth, dist);
    vec3 baseColor = colDiffuse.rgb * fragColor.rgb;
    vec3 lineColor = vec3(0.36, 0.28, 0.18);
    vec3 color = mix(baseColor, lineColor, lineMask);

    finalColor = vec4(color, colDiffuse.a * fragColor.a);
}
