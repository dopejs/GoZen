---
title: ロードバランシング
---

# ロードバランシング

GoZen は基本的な failover 以外にも複数のプロバイダー選択戦略をサポートします。プロファイルごとに戦略を選び、ヘルスチェックと組み合わせ、可用性、レイテンシ、コストに基づいてトラフィックを制御できます。

## 利用可能な戦略

### Failover

成功するまで順番にプロバイダーを試します。デフォルトの戦略で、プライマリ/バックアップ構成に適しています。

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### Round robin

複数の同等なプロバイダーにリクエストを均等に分散します。

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### Least latency

直近の応答時間が最も低いプロバイダーを優先します。

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### Least cost

要求されたモデルに対して最も安いプロバイダーを優先します。

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## ヘルス対応ルーティング

すべての戦略はヘルスモニタリングと組み合わせて使えます。`health_aware` を有効にすると、不健全なプロバイダーは復旧するまで自動的にスキップされます。

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## 戦略の選び方

- 信頼性を優先するなら `failover`
- プロバイダーがほぼ同等なら `round-robin`
- 対話的または時間に敏感な処理なら `least-latency`
- 速度よりコストを重視するなら `least-cost`

## 関連ドキュメント

- [Profiles](/docs/profiles) ではプロバイダーグループの定義方法を説明しています。
- [Routing](/docs/routing) ではシナリオベースのプロバイダー選択を扱います。
- [ヘルスモニタリング](/docs/health-monitoring) ではヘルスチェックがルーティングに与える影響を説明しています。
