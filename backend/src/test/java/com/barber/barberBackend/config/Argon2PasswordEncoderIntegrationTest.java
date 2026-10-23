package com.barber.barberBackend.config;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.security.crypto.argon2.Argon2PasswordEncoder;
import org.springframework.test.context.ActiveProfiles;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
class Argon2PasswordEncoderIntegrationTest {

    @Autowired
    private Argon2PasswordEncoder passwordEncoder;

    @Test
    void encode_and_match_withCorrectPassword_returnsTrue() {
        String raw = "miPasswordSegura123!";
        String hash = passwordEncoder.encode(raw);
        assertTrue(passwordEncoder.matches(raw, hash));
    }

    @Test
    void encode_and_match_withWrongPassword_returnsFalse() {
        String hash = passwordEncoder.encode("passwordReal");
        assertFalse(passwordEncoder.matches("passwordIncorrecta", hash));
    }

    @Test
    void encode_producesDifferentHashesEachTime() {
        String raw = "mismaPassword";
        String hash1 = passwordEncoder.encode(raw);
        String hash2 = passwordEncoder.encode(raw);
        assertNotEquals(hash1, hash2);
    }

    @Test
    void argon2HashFormat_containsExpectedMarkers() {
        String hash = passwordEncoder.encode("testPassword");
        assertTrue(hash.startsWith("$argon2id$"));
    }
}
