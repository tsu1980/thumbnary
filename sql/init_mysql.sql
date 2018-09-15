USE `thumbnary`;

--
-- Table structure for table `origin`
--

DROP TABLE IF EXISTS `origin`;
CREATE TABLE `origin` (
  `ID` char(8) NOT NULL COMMENT '8 bytes random string composed of alphanumeric characters. Use as subdomain(No digits only name)',
  `SourceType` int(11) NOT NULL COMMENT 'Source type(1=http)',
  `Scheme` char(10) NOT NULL COMMENT 'Scheme(http or https)',
  `Host` char(255) NOT NULL COMMENT 'Hostname',
  `PathPrefix` char(255) NOT NULL COMMENT 'Path prefix(starts with "/")',
  `URLSignatureKeyEnabled` tinyint(1) NOT NULL COMMENT 'URL signature is require',
  `URLSignatureKey` char(43) NOT NULL COMMENT 'URL signature key(32 bytes base64url string)',
  `URLSignatureKey_Previous` char(43) NOT NULL COMMENT 'Previous URL signature key',
  `URLSignatureKey_Version` int(11) unsigned NOT NULL COMMENT 'URL signature key version(1 or larger)',
  `CreatedDateJST` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `LastUpdatedDateJST` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`ID`),
  KEY `idx_LastUpdatedDateJST` (`LastUpdatedDateJST`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
