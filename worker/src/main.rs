mod compile;
mod db;
mod s3;

use std::env;
use std::time::Duration;

use sqlx::postgres::PgPool;

#[tokio::main]
async fn main() {
    dotenvy::dotenv().ok();
    env_logger::init();

    let pool = PgPool::connect(&env::var("DBSTRING").expect("DBSTRING not set"))
        .await
        .expect("failed to connect to database");

    let s3 = s3::S3::new(
        &env::var("S3_ENDPOINT").expect("S3_ENDPOINT not set"),
        &env::var("S3_REGION").expect("S3_REGION not set"),
        &env::var("S3_ACCESS_KEY").expect("S3_ACCESS_KEY not set"),
        &env::var("S3_SECRET_KEY").expect("S3_SECRET_KEY not set"),
        &env::var("S3_BUCKET").expect("S3_BUCKET not set"),
    )
    .await;

    let poll_interval = env::var("POLL_INTERVAL_SECS")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(10);

    log::info!("worker started, polling every {poll_interval}s");

    loop {
        match db::fetch_pending(&pool).await {
            Ok(Some(article)) => {
                log::info!(
                    "processing article id={} name={}",
                    article.id,
                    article.document_name
                );

                if let Err(e) = process_article(&pool, &s3, &article).await {
                    log::error!("failed to process article id={}: {e}", article.id);
                }
            }
            Ok(None) => {
                tokio::time::sleep(Duration::from_secs(poll_interval)).await;
            }
            Err(e) => {
                log::error!("failed to fetch pending articles: {e}");
                tokio::time::sleep(Duration::from_secs(poll_interval)).await;
            }
        }
    }
}

async fn process_article(
    pool: &PgPool,
    s3: &s3::S3,
    article: &db::Article,
) -> Result<(), Box<dyn std::error::Error>> {
    let source_bytes = s3.download(&article.document_fid).await?;
    let source = String::from_utf8(source_bytes)
        .map_err(|e| format!("invalid utf8 in document: {e}"))?;

    match compile::compile(&source) {
        Ok(pdf_bytes) => {
            let pdf_key = format!("{}.pdf", &article.document_fid);

            s3.upload(&pdf_key, pdf_bytes).await?;

            db::mark_published(pool, article.id, &pdf_key).await?;

            log::info!("article id={} published", article.id);
        }
        Err(e) => {
            log::error!("compilation failed for article id={}: {e}", article.id);
            db::mark_failed(pool, article.id).await?;
        }
    }

    Ok(())
}
