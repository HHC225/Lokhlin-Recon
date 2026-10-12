# Recon 수집 서비스

프로젝트, 자산, 하위 도메인, URL, 스캔 작업, 템플릿, 노드와 설정을 관리하는
내부 수집 서비스입니다. Vue 화면, Go REST·MCP 서버, MongoDB, Redis와 별도 스캐너를 사용합니다.
최종 제품명은 콘솔의 표시 이름 설정에서 관리합니다.

콘솔 관리자의 수집 탭에서 기존 로그인 세션으로 전체 화면을 사용합니다.
서버와 MCP는 내부 네트워크에 있으며, 콘솔 게이트웨이가 60초짜리 서명된 사용자 정보를 전달합니다.

## 소스로 이미지 만들기

```sh
docker build -f dockerfile -t recon-server:local .
```

Node 22로 `frontend`를 빌드한 결과를 `cmd/main/static`에 넣고 `./cmd/main`을 Go로 컴파일합니다.
완성된 이미지의 실행 파일은 `recon-server`이며, 실행 중 소스 저장소는 필요하지 않습니다.
`RECON_STATE_DIR=/opt/recon` 볼륨에 설정, 키와 파일을 저장합니다.

콘솔 저장소의 [로컬 수집 서비스 안내](../../artext/docs/recon-local.md)에는
스캐너 이미지 빌드, 비밀키 생성, 기존 콘솔 Compose에 연결하는 명령이 있습니다.
형제 저장소 위치가 다르면 `RECON_SOURCE_DIR`와 `RECON_WORKER_SOURCE_DIR`를 지정합니다.

## 서버 연결 설정

- `RECON_GATEWAY_SECRET`: 콘솔과 공유하는 최소 32바이트 비밀값.
- `MONGODB_IP`, `MONGODB_PORT`, `MONGODB_DATABASE`, `MONGODB_USER`, `MONGODB_PASSWORD`.
- `REDIS_IP`, `REDIS_PORT`, `REDIS_PASSWORD`.
- 새 저장소의 DB 기본 이름은 `Recon`; 기존 데이터에는 실제 DB 이름을 지정합니다.

수집 서버와 스캐너는 같은 MongoDB와 Redis에 연결해야 합니다.
`single-host-deployment.yml`은 네 서비스를 소스로 빌드하며 호스트 포트를 공개하지 않습니다.
콘솔에서 접근하려면 콘솔의 `docker-compose.recon.yml` 오버레이를 사용합니다.

## 수집 흐름

프로젝트의 수집 범위를 등록하고, 모듈과 플러그인으로 스캔 템플릿을 구성한 뒤
온라인 노드를 선택해 작업을 만듭니다. 결과는 프로젝트·자산·작업 화면에 모입니다.
하위 도메인 열거, 하위 도메인 탈취 탐지, 포트·서비스 식별, 디렉터리 탐색, 취약점 템플릿,
민감 정보 탐지, 크롤링, 페이지 모니터링과 사용자 플러그인을 지원합니다.

MCP 사용법은 [SKILL.md](SKILL.md)를 확인하세요.
