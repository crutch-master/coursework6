use sqlx::FromRow;
use sqlx::postgres::PgPool;

#[derive(FromRow)]
pub struct Article {
    pub id: i64,
    pub document_name: String,
    pub document_fid: String,
}

pub async fn fetch_pending(pool: &PgPool) -> Result<Option<Article>, sqlx::Error> {
    sqlx::query_as::<_, Article>(
        "SELECT id, document_name, document_fid FROM articles WHERE status = 'pending' ORDER BY id ASC LIMIT 1",
    )
    .fetch_optional(pool)
    .await
}

pub async fn mark_published(pool: &PgPool, id: i64, pdf_fid: &str) -> Result<(), sqlx::Error> {
    sqlx::query(
        "UPDATE articles SET status = 'published', pdf_fid = $1 WHERE id = $2",
    )
    .bind(pdf_fid)
    .bind(id)
    .execute(pool)
    .await?;
    Ok(())
}

pub async fn mark_failed(pool: &PgPool, id: i64) -> Result<(), sqlx::Error> {
    sqlx::query("UPDATE articles SET status = 'failed' WHERE id = $1")
        .bind(id)
        .execute(pool)
        .await?;
    Ok(())
}
