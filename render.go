package main

import (
	"fmt"
	"math"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ===================== SHADERE =====================

const litVS = `#version 330
in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec3 vertexNormal;
in vec4 vertexColor;
uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 matNormal;
uniform mat4 lightVP;
out vec3 fragPos;
out vec3 fragNormal;
out vec4 fragColor;
out vec2 fragData;
out vec4 fragLight;
void main() {
    vec4 wp = matModel * vec4(vertexPosition, 1.0);
    fragPos = wp.xyz;
    fragNormal = normalize((matNormal * vec4(vertexNormal, 0.0)).xyz);
    fragColor = vertexColor;
    fragData = vertexTexCoord;
    fragLight = lightVP * wp;
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}`

const litFS = `#version 330
in vec3 fragPos;
in vec3 fragNormal;
in vec4 fragColor;
in vec2 fragData;
in vec4 fragLight;
uniform vec4 colDiffuse;
uniform vec3 sunDir;
uniform vec3 sunCol;
uniform vec3 skyCol;
uniform vec3 groundCol;
uniform vec3 viewPos;
uniform vec3 fogCol;
uniform float fogDensity;
uniform sampler2D shadowMap;
uniform float shadowOn;
uniform float shadowTexel;
uniform float flash;
out vec4 finalColor;

float shadowAt(vec3 n, float ndl) {
    if (shadowOn < 0.5) return 1.0;
    vec3 p = fragLight.xyz / fragLight.w;
    p = p * 0.5 + 0.5;
    if (p.x <= 0.0 || p.x >= 1.0 || p.y <= 0.0 || p.y >= 1.0 || p.z >= 1.0) return 1.0;
    float bias = max(0.0022 * (1.0 - ndl), 0.0005);
    float s = 0.0;
    for (int x = -1; x <= 1; x++) {
        for (int y = -1; y <= 1; y++) {
            float d = texture(shadowMap, p.xy + vec2(x, y) * shadowTexel).r;
            s += (p.z - bias > d) ? 0.0 : 1.0;
        }
    }
    s /= 9.0;
    vec2 e = min(p.xy, 1.0 - p.xy);
    float fade = clamp(min(e.x, e.y) * 14.0, 0.0, 1.0);
    return mix(1.0, s, fade);
}

void main() {
    vec4 base = fragColor * colDiffuse;
    vec3 n = normalize(fragNormal);
    if (!gl_FrontFacing) n = -n;
    float ao = fragData.x;
    float emis = fragData.y;
    float ndl = dot(n, sunDir);
    float diff = max(ndl, 0.0);
    float sh = (ndl > 0.0) ? shadowAt(n, ndl) : 0.0;
    vec3 hemi = mix(groundCol, skyCol, n.y * 0.5 + 0.5);
    vec3 v = normalize(viewPos - fragPos);
    vec3 h = normalize(sunDir + v);
    float spec = pow(max(dot(n, h), 0.0), 42.0) * 0.20 * sh * step(0.0, ndl);
    float rim = pow(1.0 - max(dot(n, v), 0.0), 3.0) * 0.16;
    vec3 fillDir = normalize(vec3(-sunDir.x, 0.35, -sunDir.z));
    float fill = max(dot(n, fillDir), 0.0) * 0.14;
    // suprafetele din tuneluri (ao < 0.3) nu primesc soare, chiar si departe de harta umbrelor
    float sunOK = step(0.3, ao);
    vec3 lit = base.rgb * (hemi * 0.66 * ao + sunCol * diff * sh * 0.84 * mix(0.35, 1.0, ao) * sunOK + skyCol * fill * ao) + sunCol * spec * sunOK + skyCol * rim * ao;
    lit = mix(lit, base.rgb * 1.1, emis);
    lit += vec3(flash);
    // umar moale pentru lumini puternice (nu mai "arde" in alb)
    vec3 hi = max(lit - 0.72, 0.0);
    lit = min(lit, 0.72) + 0.28 * (1.0 - exp(-hi / 0.28));
    float dist = length(viewPos - fragPos);
    float fog = 1.0 - exp(-pow(dist * fogDensity, 1.7));
    lit = mix(lit, fogCol, clamp(fog, 0.0, 1.0));
    finalColor = vec4(lit, base.a);
}`

const depthVS = `#version 330
in vec3 vertexPosition;
uniform mat4 mvp;
void main() { gl_Position = mvp * vec4(vertexPosition, 1.0); }`

const depthFS = `#version 330
out vec4 finalColor;
void main() { finalColor = vec4(1.0); }`

const skyVS = `#version 330
in vec3 vertexPosition;
uniform mat4 mvp;
out vec3 dir;
void main() {
    dir = vertexPosition;
    vec4 p = mvp * vec4(vertexPosition, 1.0);
    gl_Position = p.xyww;
}`

const skyFS = `#version 330
in vec3 dir;
uniform vec3 sunDir;
uniform vec3 zenithCol;
uniform vec3 horizonCol;
uniform vec3 sunCol;
out vec4 finalColor;
void main() {
    vec3 d = normalize(dir);
    float t = clamp(d.y, 0.0, 1.0);
    vec3 col = mix(horizonCol, zenithCol, pow(t, 0.5));
    if (d.y < 0.0) col = mix(horizonCol, horizonCol * 0.92, clamp(-d.y * 4.0, 0.0, 1.0));
    float sd = max(dot(d, sunDir), 0.0);
    col += sunCol * (pow(sd, 1200.0) * 3.0 + pow(sd, 60.0) * 0.35 + pow(sd, 8.0) * 0.12);
    finalColor = vec4(col, 1.0);
}`

const waterVS = `#version 330
in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec4 vertexColor;
uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 lightVP;
uniform float time;
out vec3 fragPos;
out vec3 fragNormal;
out vec2 fragData;
out vec4 fragLight;
void main() {
    vec3 p = vertexPosition;
    float depth = vertexTexCoord.x;
    float amp = clamp(depth / 4.0, 0.15, 1.0);
    float w1 = sin(p.x * 0.11 + time * 1.3) * 0.22;
    float w2 = sin(p.z * 0.13 + time * 1.1 + p.x * 0.05) * 0.18;
    float w3 = sin((p.x + p.z) * 0.31 + time * 2.4) * 0.06;
    p.y += (w1 + w2 + w3) * amp;
    float dx = (cos(p.x * 0.11 + time * 1.3) * 0.22 * 0.11 + cos(p.z * 0.13 + time * 1.1 + p.x * 0.05) * 0.18 * 0.05 + cos((p.x + p.z) * 0.31 + time * 2.4) * 0.06 * 0.31) * amp;
    float dz = (cos(p.z * 0.13 + time * 1.1 + p.x * 0.05) * 0.18 * 0.13 + cos((p.x + p.z) * 0.31 + time * 2.4) * 0.06 * 0.31) * amp;
    fragNormal = normalize(vec3(-dx, 1.0, -dz));
    vec4 wp = matModel * vec4(p, 1.0);
    fragPos = wp.xyz;
    fragData = vertexTexCoord;
    fragLight = lightVP * wp;
    gl_Position = mvp * vec4(p, 1.0);
}`

const waterFS = `#version 330
in vec3 fragPos;
in vec3 fragNormal;
in vec2 fragData;
in vec4 fragLight;
uniform vec3 sunDir;
uniform vec3 sunCol;
uniform vec3 skyCol;
uniform vec3 viewPos;
uniform vec3 fogCol;
uniform float fogDensity;
uniform float time;
uniform sampler2D shadowMap;
uniform float shadowOn;
uniform float shadowTexel;
out vec4 finalColor;
float shadowAt() {
    if (shadowOn < 0.5) return 1.0;
    vec3 p = fragLight.xyz / fragLight.w;
    p = p * 0.5 + 0.5;
    if (p.x <= 0.0 || p.x >= 1.0 || p.y <= 0.0 || p.y >= 1.0 || p.z >= 1.0) return 1.0;
    float s = 0.0;
    for (int x = -1; x <= 1; x++) for (int y = -1; y <= 1; y++) {
        float d = texture(shadowMap, p.xy + vec2(x, y) * shadowTexel * 1.5).r;
        s += (p.z - 0.002 > d) ? 0.0 : 1.0;
    }
    return mix(0.55, 1.0, s / 9.0);
}
void main() {
    float depth = fragData.x;
    vec3 n = normalize(fragNormal);
    vec3 v = normalize(viewPos - fragPos);
    vec3 shallow = vec3(0.30, 0.86, 0.84);
    vec3 deep = vec3(0.07, 0.38, 0.70);
    vec3 col = mix(shallow, deep, smoothstep(0.2, 7.0, depth));
    float fres = pow(1.0 - max(dot(n, v), 0.0), 4.0);
    col = mix(col, skyCol * 1.05, fres * 0.55);
    float sh = shadowAt();
    col *= mix(0.75, 1.0, sh);
    vec3 r = reflect(-sunDir, n);
    float spec = pow(max(dot(r, v), 0.0), 160.0) * 1.6 * sh;
    col += sunCol * spec;
    float ripple = sin(fragPos.x * 1.7 + time * 2.0) * sin(fragPos.z * 1.9 - time * 1.6);
    float foam = (1.0 - smoothstep(0.0, 1.1, depth)) * (0.65 + 0.35 * ripple);
    col = mix(col, vec3(0.97, 0.99, 1.0), clamp(foam, 0.0, 1.0) * 0.85);
    float alpha = mix(0.55, 0.93, smoothstep(0.0, 3.5, depth));
    alpha = max(alpha, foam * 0.9);
    float dist = length(viewPos - fragPos);
    float fog = 1.0 - exp(-pow(dist * fogDensity, 1.7));
    col = mix(col, fogCol, clamp(fog, 0.0, 1.0));
    finalColor = vec4(col, alpha);
}`

// ===================== RENDERER =====================

type Chunk struct {
	Mesh    rl.Mesh
	Center  rl.Vector3
	Radius  float32
	Shadows bool
}

type Renderer struct {
	Lit      rl.Shader
	Depth    rl.Shader
	Sky      rl.Shader
	Water    rl.Shader
	MatLit   rl.Material
	MatDepth rl.Material
	MatSky   rl.Material
	MatWater rl.Material

	SkyMesh rl.Mesh

	ShadowRT    rl.RenderTexture2D
	ShadowSize  int32
	ShadowsOn   bool
	LightVP     rl.Matrix
	ShadowRange float32

	SunDir     rl.Vector3
	SunCol     [3]float32
	SkyCol     [3]float32
	GroundCol  [3]float32
	FogCol     [3]float32
	ZenithCol  [3]float32
	HorizonCol [3]float32
	FogDensity float32

	locLit   map[string]int32
	locWater map[string]int32
	locSky   map[string]int32

	frustum [6][4]float32
	camPos  rl.Vector3
	time    float32
}

func newShaderLocs(sh rl.Shader, names ...string) map[string]int32 {
	m := map[string]int32{}
	for _, n := range names {
		m[n] = rl.GetShaderLocation(sh, n)
	}
	return m
}

func NewRenderer() *Renderer {
	r := &Renderer{ShadowSize: 2048, ShadowsOn: true, ShadowRange: 90}

	dbg := func(m string) {
		if os.Getenv("BATTLE_DEBUG") != "" {
			fmt.Println("render:", m)
		}
	}
	dbg("shader lit")
	r.Lit = rl.LoadShaderFromMemory(litVS, litFS)
	r.Lit.UpdateLocation(rl.ShaderLocMatrixModel, rl.GetShaderLocation(r.Lit, "matModel"))
	r.Lit.UpdateLocation(rl.ShaderLocMatrixNormal, rl.GetShaderLocation(r.Lit, "matNormal"))
	r.locLit = newShaderLocs(r.Lit, "sunDir", "sunCol", "skyCol", "groundCol", "viewPos", "fogCol", "fogDensity", "shadowMap", "shadowOn", "shadowTexel", "lightVP", "flash")

	dbg("shader depth")
	r.Depth = rl.LoadShaderFromMemory(depthVS, depthFS)
	r.Sky = rl.LoadShaderFromMemory(skyVS, skyFS)
	r.locSky = newShaderLocs(r.Sky, "sunDir", "zenithCol", "horizonCol", "sunCol")

	dbg("shader apa")
	r.Water = rl.LoadShaderFromMemory(waterVS, waterFS)
	r.Water.UpdateLocation(rl.ShaderLocMatrixModel, rl.GetShaderLocation(r.Water, "matModel"))
	r.locWater = newShaderLocs(r.Water, "sunDir", "sunCol", "skyCol", "viewPos", "fogCol", "fogDensity", "time", "shadowMap", "shadowOn", "shadowTexel", "lightVP")

	r.MatLit = rl.LoadMaterialDefault()
	r.MatLit.Shader = r.Lit
	r.MatDepth = rl.LoadMaterialDefault()
	r.MatDepth.Shader = r.Depth
	r.MatSky = rl.LoadMaterialDefault()
	r.MatSky.Shader = r.Sky
	r.MatWater = rl.LoadMaterialDefault()
	r.MatWater.Shader = r.Water

	dbg("materiale ok")
	sb := NewMB()
	sb.Ellipsoid(v3(0, 0, 0), v3(1, 1, 1), 16, 24, rl.White)
	r.SkyMesh = sb.Build()

	// iluminare de zi, calda (mereu zi - regula de produs)
	r.SunDir = rl.Vector3Normalize(v3(0.45, 0.78, 0.38))
	r.SunCol = [3]float32{1.0, 0.94, 0.82}
	r.SkyCol = [3]float32{0.56, 0.68, 0.86}
	r.GroundCol = [3]float32{0.42, 0.40, 0.33}
	r.FogCol = [3]float32{0.74, 0.86, 0.97}
	r.ZenithCol = [3]float32{0.25, 0.53, 0.92}
	r.HorizonCol = [3]float32{0.76, 0.88, 0.98}
	r.FogDensity = 0.0016

	dbg("cer ok")
	r.createShadowMap()
	dbg("umbre ok")
	r.applyStaticUniforms()
	dbg("uniforme ok")
	return r
}

func (r *Renderer) createShadowMap() {
	sz := r.ShadowSize
	rt := rl.RenderTexture2D{}
	rt.ID = rl.LoadFramebuffer()
	rt.Texture.Width = sz
	rt.Texture.Height = sz
	if rt.ID > 0 {
		rl.EnableFramebuffer(rt.ID)
		rt.Depth.ID = rl.LoadTextureDepth(sz, sz, false)
		rt.Depth.Width = sz
		rt.Depth.Height = sz
		rt.Depth.Format = 19
		rt.Depth.Mipmaps = 1
		rl.FramebufferAttach(rt.ID, rt.Depth.ID, rl.AttachmentDepth, rl.AttachmentTexture2d, 0)
		if !rl.FramebufferComplete(rt.ID) {
			r.ShadowsOn = false
		}
		rl.DisableFramebuffer()
	} else {
		r.ShadowsOn = false
	}
	r.ShadowRT = rt
}

func setV3(sh rl.Shader, loc int32, v [3]float32) {
	if loc >= 0 {
		rl.SetShaderValue(sh, loc, v[:], rl.ShaderUniformVec3)
	}
}

func setF(sh rl.Shader, loc int32, f float32) {
	if loc >= 0 {
		rl.SetShaderValue(sh, loc, []float32{f}, rl.ShaderUniformFloat)
	}
}

func (r *Renderer) applyStaticUniforms() {
	sd := [3]float32{r.SunDir.X, r.SunDir.Y, r.SunDir.Z}
	setV3(r.Lit, r.locLit["sunDir"], sd)
	setV3(r.Lit, r.locLit["sunCol"], r.SunCol)
	setV3(r.Lit, r.locLit["skyCol"], r.SkyCol)
	setV3(r.Lit, r.locLit["groundCol"], r.GroundCol)
	setV3(r.Lit, r.locLit["fogCol"], r.FogCol)
	setF(r.Lit, r.locLit["fogDensity"], r.FogDensity)
	setF(r.Lit, r.locLit["shadowTexel"], 1.0/float32(r.ShadowSize))
	setF(r.Lit, r.locLit["flash"], 0)

	setV3(r.Water, r.locWater["sunDir"], sd)
	setV3(r.Water, r.locWater["sunCol"], r.SunCol)
	setV3(r.Water, r.locWater["skyCol"], r.HorizonCol)
	setV3(r.Water, r.locWater["fogCol"], r.FogCol)
	setF(r.Water, r.locWater["fogDensity"], r.FogDensity)
	setF(r.Water, r.locWater["shadowTexel"], 1.0/float32(r.ShadowSize))

	setV3(r.Sky, r.locSky["sunDir"], sd)
	setV3(r.Sky, r.locSky["zenithCol"], r.ZenithCol)
	setV3(r.Sky, r.locSky["horizonCol"], r.HorizonCol)
	setV3(r.Sky, r.locSky["sunCol"], r.SunCol)

	slot := []int32{10}
	if loc := r.locLit["shadowMap"]; loc >= 0 {
		rl.EnableShader(r.Lit.ID)
		rl.SetUniform(loc, slot, int32(rl.ShaderUniformInt), 1)
		rl.DisableShader()
	}
	if loc := r.locWater["shadowMap"]; loc >= 0 {
		rl.EnableShader(r.Water.ID)
		rl.SetUniform(loc, slot, int32(rl.ShaderUniformInt), 1)
		rl.DisableShader()
	}
}

// BeginFrame pregateste uniformele care depind de camera si timp.
func (r *Renderer) BeginFrame(cam rl.Camera3D, t float32) {
	r.camPos = cam.Position
	r.time = t
	vp := [3]float32{cam.Position.X, cam.Position.Y, cam.Position.Z}
	setV3(r.Lit, r.locLit["viewPos"], vp)
	setV3(r.Water, r.locWater["viewPos"], vp)
	setF(r.Water, r.locWater["time"], t)
	on := float32(0)
	if r.ShadowsOn {
		on = 1
	}
	setF(r.Lit, r.locLit["shadowOn"], on)
	setF(r.Water, r.locWater["shadowOn"], on)

	aspect := float32(rl.GetScreenWidth()) / float32(maxi32(int32(rl.GetScreenHeight()), 1))
	view := rl.MatrixLookAt(cam.Position, cam.Target, cam.Up)
	proj := rl.MatrixPerspective(cam.Fovy*rl.Deg2rad, aspect, 0.05, 1400)
	r.extractFrustum(rl.MatrixMultiply(view, proj))
}

func maxi32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func (r *Renderer) extractFrustum(m rl.Matrix) {
	// planele frustum-ului din matricea view*proj (conventie raylib, coloane)
	rows := [4][4]float32{
		{m.M0, m.M4, m.M8, m.M12},
		{m.M1, m.M5, m.M9, m.M13},
		{m.M2, m.M6, m.M10, m.M14},
		{m.M3, m.M7, m.M11, m.M15},
	}
	set := func(i int, a, b [4]float32, sign float32) {
		p := [4]float32{a[0] + sign*b[0], a[1] + sign*b[1], a[2] + sign*b[2], a[3] + sign*b[3]}
		l := sqrtf(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
		if l > 0 {
			p[0] /= l
			p[1] /= l
			p[2] /= l
			p[3] /= l
		}
		r.frustum[i] = p
	}
	set(0, rows[3], rows[0], 1)
	set(1, rows[3], rows[0], -1)
	set(2, rows[3], rows[1], 1)
	set(3, rows[3], rows[1], -1)
	set(4, rows[3], rows[2], 1)
	set(5, rows[3], rows[2], -1)
}

func (r *Renderer) SphereVisible(c rl.Vector3, rad float32) bool {
	for i := 0; i < 6; i++ {
		p := r.frustum[i]
		if p[0]*c.X+p[1]*c.Y+p[2]*c.Z+p[3] < -rad {
			return false
		}
	}
	return true
}

// ShadowPass deseneaza adancimea vazuta din soare, intr-o zona patrata in jurul jucatorului.
func (r *Renderer) ShadowPass(focus rl.Vector3, draw func(mat rl.Material)) {
	if !r.ShadowsOn {
		return
	}
	rng := r.ShadowRange
	texel := (2 * rng) / float32(r.ShadowSize)
	// stabilizare: mutam centrul doar in pasi de un texel (umbrele nu mai "tremura")
	right := rl.Vector3Normalize(rl.Vector3CrossProduct(v3(0, 1, 0), r.SunDir))
	up := rl.Vector3CrossProduct(r.SunDir, right)
	fx := rl.Vector3DotProduct(focus, right)
	fy := rl.Vector3DotProduct(focus, up)
	fx = float32(math.Floor(float64(fx/texel))) * texel
	fy = float32(math.Floor(float64(fy/texel))) * texel
	fz := rl.Vector3DotProduct(focus, r.SunDir)
	center := rl.Vector3Add(rl.Vector3Add(rl.Vector3Scale(right, fx), rl.Vector3Scale(up, fy)), rl.Vector3Scale(r.SunDir, fz))

	eye := rl.Vector3Add(center, rl.Vector3Scale(r.SunDir, 220))
	view := rl.MatrixLookAt(eye, center, v3(0, 1, 0))
	proj := rl.MatrixOrtho(-rng, rng, -rng, rng, 1, 480)
	r.LightVP = rl.MatrixMultiply(view, proj)

	// dezlegam harta de umbre din slotul 10 cat timp scriem in ea
	rl.ActiveTextureSlot(10)
	rl.DisableTexture()
	rl.ActiveTextureSlot(0)

	rl.BeginTextureMode(r.ShadowRT)
	rl.ClearBackground(rl.White)
	rl.DrawRenderBatchActive()
	rl.SetMatrixProjection(proj)
	rl.SetMatrixModelview(view)
	rl.EnableDepthTest()
	rl.DisableBackfaceCulling()
	draw(r.MatDepth)
	rl.EnableBackfaceCulling()
	rl.DrawRenderBatchActive()
	rl.EndTextureMode()

	if loc := r.locLit["lightVP"]; loc >= 0 {
		rl.SetShaderValueMatrix(r.Lit, loc, r.LightVP)
	}
	if loc := r.locWater["lightVP"]; loc >= 0 {
		rl.SetShaderValueMatrix(r.Water, loc, r.LightVP)
	}
}

// BindShadow leaga harta de umbre in slotul 10, inainte de pasul principal.
func (r *Renderer) BindShadow() {
	if !r.ShadowsOn {
		return
	}
	rl.ActiveTextureSlot(10)
	rl.EnableTexture(r.ShadowRT.Depth.ID)
	rl.ActiveTextureSlot(0)
}

func (r *Renderer) InShadowRange(c rl.Vector3, rad float32, focus rl.Vector3) bool {
	dx := c.X - focus.X
	dz := c.Z - focus.Z
	lim := r.ShadowRange*1.5 + rad
	return dx*dx+dz*dz < lim*lim
}

func (r *Renderer) DrawSky(cam rl.Camera3D) {
	rl.DisableDepthMask()
	rl.DisableBackfaceCulling()
	m := rl.MatrixMultiply(rl.MatrixScale(900, 900, 900), rl.MatrixTranslate(cam.Position.X, cam.Position.Y, cam.Position.Z))
	rl.DrawMesh(r.SkyMesh, r.MatSky, m)
	rl.EnableBackfaceCulling()
	rl.EnableDepthMask()
}

// DrawLit deseneaza un mesh cu iluminare; tint inmulteste culoarea (alpha pentru fantome).
func (r *Renderer) DrawLit(mesh rl.Mesh, m rl.Matrix, tint rl.Color) {
	r.MatLit.Maps.Color = tint
	rl.DrawMesh(mesh, r.MatLit, m)
}

func (r *Renderer) SetFlash(f float32) {
	setF(r.Lit, r.locLit["flash"], f)
}

// SetIndoor: in tunel ceata e intunecata si deasa (capatul tunelului se pierde in umbra calda).
func (r *Renderer) SetIndoor(k float32) {
	fc := [3]float32{
		r.FogCol[0] + (0.16-r.FogCol[0])*k,
		r.FogCol[1] + (0.15-r.FogCol[1])*k,
		r.FogCol[2] + (0.13-r.FogCol[2])*k,
	}
	d := r.FogDensity + (0.0055-r.FogDensity)*k
	setV3(r.Lit, r.locLit["fogCol"], fc)
	setF(r.Lit, r.locLit["fogDensity"], d)
}
