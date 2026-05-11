use aws_config::{BehaviorVersion, Region, defaults};
use aws_credential_types::Credentials;
use aws_sdk_s3::Client;
use aws_sdk_s3::config::RequestChecksumCalculation;

pub struct S3 {
    client: Client,
    bucket: String,
}

impl S3 {
    pub async fn new(
        endpoint: &str,
        region: &str,
        access_key: &str,
        secret_key: &str,
        bucket: &str,
    ) -> Self {
        let creds = Credentials::new(access_key, secret_key, None, None, "env");

        let config = defaults(BehaviorVersion::latest())
            .region(Region::new(region.to_string()))
            .credentials_provider(creds)
            .endpoint_url(endpoint)
            .load()
            .await;

        let client = Client::from_conf(
            aws_sdk_s3::Config::from(&config)
                .to_builder()
                .force_path_style(true)
                .request_checksum_calculation(RequestChecksumCalculation::WhenRequired)
                .build(),
        );

        let s3 = Self {
            client,
            bucket: bucket.to_string(),
        };

        s3.ensure_bucket().await;
        s3
    }

    async fn ensure_bucket(&self) {
        let exists = self
            .client
            .head_bucket()
            .bucket(&self.bucket)
            .send()
            .await
            .is_ok();

        if !exists {
            match self
                .client
                .create_bucket()
                .bucket(&self.bucket)
                .send()
                .await
            {
                Ok(_) => log::info!("created bucket: {}", self.bucket),
                Err(e) => log::error!("failed to create bucket: {e}"),
            }
        }
    }

    pub async fn download(&self, key: &str) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
        let resp = self
            .client
            .get_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await?;

        let bytes = resp.body.collect().await?.to_vec();
        Ok(bytes)
    }

    pub async fn upload(
        &self,
        key: &str,
        data: Vec<u8>,
    ) -> Result<(), Box<dyn std::error::Error>> {
        self.client
            .put_object()
            .bucket(&self.bucket)
            .key(key)
            .body(data.into())
            .send()
            .await?;
        Ok(())
    }
}
