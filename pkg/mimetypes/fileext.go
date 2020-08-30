package mimetypes

import (
	"strings"
)

// ContentTypeDefaultFileExt return default file extension for content type
func ContentTypeDefaultFileExt(contentType string) string {
	if strings.HasPrefix(contentType, "text/") {
		switch contentType {
		case "text/cmd":
			return ".cmd"
		case "text/css":
			return ".css"
		case "text/csv":
			return ".csv"
		case "text/javascript":
			return ".js"
		case "text/plain":
			return ".txt"
		case "text/php":
			return ".php"
		case "text/xml":
			return ".xml"
		case "text/markdown":
			return ".md"
		case "text/html":
			return ".html"
		}
	} else if strings.HasPrefix(contentType, "image/") {
		switch contentType {
		case "image/gif":
			return ".gif"
		case "image/jpeg", "image/pjpeg":
			return ".jpg"
		case "image/svg+xml":
			return ".svg"
		case "image/tiff":
			return ".tiff"
		case "image/vnd.microsoft.icon":
			return ".ico"
		case "image/vnd.wap.wbmp":
			return ".vnd.wap.wbmp"
		case "image/webp":
			return ".webp"
		}
	} else if strings.HasPrefix(contentType, "application/") {
		switch contentType {
		case "application/json":
			return ".json"
		case "application/javascript":
			return ".js"
		case "application/octet-stream":
			return ""
		case "application/ogg":
			return ".ogg"
		case "application/pdf":
			return ".pdf"
		case "application/postscript":
			return ".postscript"
		case "application/soap+xml":
			return ".xml"
		case "application/font-woff":
			return ".woff"
		case "application/xhtml+xml":
			return ".xhtml"
		case "application/xml-dtd":
			return ".xml"
		case "application/zip":
			return ".zip"
		case "application/gzip":
			return ".gz"
		case "application/x-bittorrent":
			return ".torrent"
		case "application/x-rar-compressed":
			return ".rar"
		case "application/x-tex", "application/x-latex":
			return ".tex"
		case "application/x-shockwave-flash":
			return ".swf"
		case "application/x-font-ttf":
			return ".ttf"
		case "application/xml":
			return ".xml"
		case "application/msword":
			return ".doc"
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			return ".docx"
		case "application/vnd.ms-excel":
			return ".xls"
		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
			return ".xlsx"
		case "application/vnd.ms-excel.sheet.macroEnabled.12":
			return ".xlsm"
		case "application/vnd.ms-powerpoint":
			return ".ppt"
		case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
			return ".pptx"
		case "application/vnd.oasis.opendocument.text":
			return ".odt"
		case "application/vnd.oasis.opendocument.text-template":
			return ".ott"
		case "application/vnd.oasis.opendocument.graphics":
			return ".odg"
		case "application/vnd.oasis.opendocument.graphics-template":
			return ".otg"
		case "application/vnd.oasis.opendocument.presentation":
			return ".odp"
		case "application/vnd.oasis.opendocument.presentation-template":
			return ".otp"
		case "application/vnd.oasis.opendocument.spreadsheet":
			return ".ods"
		case "application/vnd.oasis.opendocument.spreadsheet-template":
			return ".ots"
		case "application/vnd.oasis.opendocument.chart":
			return ".odc"
		case "application/vnd.oasis.opendocument.chart-template":
			return ".otc"
		case "application/vnd.oasis.opendocument.image":
			return ".odi"
		case "application/vnd.oasis.opendocument.image-template":
			return ".oti"
		case "application/vnd.oasis.opendocument.formula":
			return ".odf"
		case "application/vnd.oasis.opendocument.formula-template":
			return ".otf"
		case "application/vnd.oasis.opendocument.text-master":
			return ".odm"
		case "application/vnd.oasis.opendocument.text-web":
			return ".oth"
		case "application/x-dvi":
			return ".dvi"
		case "application/x-pkcs7-certificates":
			return ".p7b"
		case "application/x-pkcs7-certreqresp":
			return ".p7r"
		case "application/x-pkcs7-mime":
			return ".p7c"
		case "application/x-pkcs7-signature":
			return ".p7s"
		case "application/vnd.google-earth.kml+xml":
			return ".kml"
		}
	} else if strings.HasPrefix(contentType, "audio/") {
		switch contentType {
		case "audio/L24":
			return ".pcm"
		case "audio/aac":
			return ".aac"
		case "audio/basic":
			return ".mulaw"
		case "audio/mp4":
			return ".mp4"
		case "audio/ogg":
			return ".ogg"
		case "audio/x-ms-wma", "audio/x-ms-wax":
			return ".wma"
		case "audio/vnd.wave":
			return ".wav"
		case "audio/vnd.rn-realaudio":
			return ".ra"
		case "audio/mpeg":
			return ".mpg"
		}
	} else if strings.HasPrefix(contentType, "model/") {
		switch contentType {
		case "model/iges":
			return ".igs"
		case "model/mesh":
			return ".mesh"
		case "model/vrml":
			return ".vrml"
		case "model/x3d+binary":
			return ".x3db"
		case "model/x3d+vrml":
			return ".x3dv"
		case "model/x3d+xml":
			return ".x3d"
		}
	} else if strings.HasPrefix(contentType, "video/") {
		switch contentType {
		case "video/3gpp":
			return ".3gp"
		case "video/3gpp2":
			return ".3g2"
		case "video/mpeg":
			return ".mpg"
		case "video/mp4":
			return ".mp4"
		case "video/ogg":
			return ".ogg"
		case "video/webm":
			return ".webm"
		case "video/x-ms-wma", "video/x-ms-wax":
			return ".wma"
		case "video/x-ms-wmv":
			return ".wmv"
		case "video/x-flv":
			return ".flv"
		}
	}
	return ""
}
