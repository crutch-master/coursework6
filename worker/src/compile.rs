use std::path::PathBuf;
use typst::LibraryExt;
use typst::foundations::{Bytes, Datetime};
use typst::layout::PagedDocument;
use typst::syntax::{FileId, Source, VirtualPath};
use typst::text::{Font, FontBook};
use typst::utils::LazyHash;
use typst::Library;
use typst::World;
use typst_pdf::PdfOptions;

struct TypstWorld {
    library: LazyHash<Library>,
    book: LazyHash<FontBook>,
    fonts: Vec<Font>,
    source: Source,
    source_id: FileId,
}

impl World for TypstWorld {
    fn library(&self) -> &LazyHash<Library> {
        &self.library
    }

    fn book(&self) -> &LazyHash<FontBook> {
        &self.book
    }

    fn main(&self) -> FileId {
        self.source_id
    }

    fn source(&self, id: FileId) -> Result<Source, typst::diag::FileError> {
        if id == self.source_id {
            Ok(self.source.clone())
        } else {
            Err(typst::diag::FileError::NotFound(PathBuf::new()))
        }
    }

    fn file(&self, _id: FileId) -> Result<Bytes, typst::diag::FileError> {
        Err(typst::diag::FileError::NotFound(PathBuf::new()))
    }

    fn font(&self, index: usize) -> Option<Font> {
        self.fonts.get(index).cloned()
    }

    fn today(&self, _offset: Option<i64>) -> Option<Datetime> {
        None
    }
}

pub fn compile(source: &str) -> Result<Vec<u8>, String> {
    let fonts: Vec<Font> = typst_assets::fonts()
        .filter_map(|data| Font::new(Bytes::new(data), 0))
        .collect();

    let book = FontBook::from_fonts(&fonts);
    let source_id = FileId::new_fake(VirtualPath::new("main.typ"));
    let source = Source::new(source_id, source.into());

    let world = TypstWorld {
        library: LazyHash::new(Library::builder().build()),
        book: LazyHash::new(book),
        fonts,
        source,
        source_id,
    };

    let doc = typst::compile::<PagedDocument>(&world)
        .output
        .map_err(|errs| format!("{errs:?}"))?;

    typst_pdf::pdf(&doc, &PdfOptions::default()).map_err(|errs| format!("{errs:?}"))
}
